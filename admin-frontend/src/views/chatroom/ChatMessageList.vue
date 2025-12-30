<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="发送用户">
          <el-input v-model="query.fromUserName" placeholder="请输入发送用户名" clearable />
        </el-form-item>
        <el-form-item label="接收用户">
          <el-input v-model="query.toUserName" placeholder="请输入接收用户名" clearable />
        </el-form-item>
        <el-form-item label="聊天室ID">
          <el-input v-model="query.roomId" placeholder="请输入聊天室ID" clearable />
        </el-form-item>
        <el-form-item label="消息类型">
          <el-select v-model="query.messageType" placeholder="请选择消息类型" clearable>
            <el-option label="文本" :value="1" />
            <el-option label="图片" :value="2" />
            <el-option label="文件" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- D2Table 组件 -->
    <el-card>
      <D2Table
        :columns="columns"
        :data="list"
        :total="total"
        :page-size="query.pageSize"
        :current-page="query.page"
        :drawer-columns="drawerColumns"
        :drawer-add-columns="[]"
        :have-edit="false"
        :have-detail="true"
        delete-permission="chat_message:delete"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
      >
        <!-- 自定义消息类型列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'messageType'" :type="getMessageTypeTagType(row.messageType)">
            {{ getMessageTypeLabel(row.messageType) }}
          </el-tag>
          <!-- 消息内容：如果是图片，显示图片预览 -->
          <div v-else-if="column.prop === 'content' && row.messageType === 2" class="message-content-image">
            <el-image
              :src="row.content"
              fit="cover"
              style="width: 100px; height: 100px"
              :preview-src-list="[row.content]"
              preview-teleported
            >
              <template #error>
                <div class="image-slot">
                  <el-icon><Picture /></el-icon>
                </div>
              </template>
            </el-image>
          </div>
          <!-- 消息内容：文本消息，显示前50个字符 -->
          <span v-else-if="column.prop === 'content' && row.messageType === 1" class="message-content-text">
            {{ row.content.length > 50 ? row.content.substring(0, 50) + '...' : row.content }}
          </span>
          <!-- 消息内容：文件消息 -->
          <el-link v-else-if="column.prop === 'content' && row.messageType === 3" :href="row.content" target="_blank" type="primary">
            查看文件
          </el-link>
          <!-- 发送时间：格式化显示 -->
          <span v-else-if="column.prop === 'createdAt'">
            {{ formatUnixTime(row.createdAt) }}
          </span>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {Picture} from '@element-plus/icons-vue';
import {chatMessageListAdmin, chatMessageDelete} from '@/api/generated/admin';
import type {ChatMessageItem, ChatMessageListReq} from '@/api/generated/admin';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';
import {buildFileUrlFromResponse} from '@/utils/file';
import {formatUnixTime} from '@/utils/date';

const query = reactive<ChatMessageListReq & {fromUserName?: string; toUserName?: string; messageType?: number}>({
  page: 1,
  pageSize: 10,
  roomId: '',
  userId: 0,
  fromUserName: '',
  toUserName: '',
  messageType: undefined
});
const list = ref<ChatMessageItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'fromUserName', label: '发送用户', width: 120},
  {prop: 'toUserName', label: '接收用户', width: 120},
  {prop: 'roomId', label: '聊天室ID', width: 150},
  {prop: 'content', label: '消息内容', minWidth: 200},
  {prop: 'messageType', label: '消息类型', width: 100},
  {prop: 'createdAt', label: '发送时间', width: 180}
]);

// 详情抽屉列配置（只读）
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'fromUserName', label: '发送用户', type: D2TableElemType.Tag},
  {prop: 'toUserName', label: '接收用户', type: D2TableElemType.Tag},
  {prop: 'roomId', label: '聊天室ID', type: D2TableElemType.Tag},
  {prop: 'content', label: '消息内容'},
  {prop: 'messageType', label: '消息类型'},
  {prop: 'createdAt', label: '发送时间', type: D2TableElemType.ConvertTime}
]);

// 获取消息类型标签
const getMessageTypeLabel = (type: number) => {
  const typeMap: Record<number, string> = {
    1: '文本',
    2: '图片',
    3: '文件'
  };
  return typeMap[type] || '未知';
};

// 获取消息类型标签颜色
const getMessageTypeTagType = (type: number): 'success' | 'warning' | 'info' => {
  const typeMap: Record<number, 'success' | 'warning' | 'info'> = {
    1: 'success',
    2: 'warning',
    3: 'info'
  };
  return typeMap[type] || 'info';
};

const loadData = async () => {
  loading.value = true;
  try {
    // 构建查询参数（后端只支持 roomId 和 userId，前端需要先查询用户ID）
    const req: ChatMessageListReq = {
      page: query.page,
      pageSize: query.pageSize,
      roomId: query.roomId || '',
      userId: query.userId || 0
    };
    
    const resp = await chatMessageListAdmin(req);
    let filteredList = resp.list || [];
    
    // 前端过滤：根据发送用户名和接收用户名过滤
    if (query.fromUserName) {
      filteredList = filteredList.filter(item => 
        item.fromUserName?.toLowerCase().includes(query.fromUserName!.toLowerCase())
      );
    }
    if (query.toUserName) {
      filteredList = filteredList.filter(item => 
        item.toUserName?.toLowerCase().includes(query.toUserName!.toLowerCase())
      );
    }
    if (query.messageType !== undefined && query.messageType !== null) {
      filteredList = filteredList.filter(item => item.messageType === query.messageType);
    }
    
    // 处理图片消息的 URL（如果是相对路径，需要拼接 baseUrl）
    filteredList = filteredList.map(item => {
      if (item.messageType === 2 && item.content && !item.content.startsWith('http')) {
        // 图片消息，如果是相对路径，使用工具函数拼接完整 URL
        // 注意：这里假设 content 存储的是文件路径，实际可能需要从文件表查询
        item.content = buildFileUrlFromResponse({path: item.content});
      }
      return item;
    });
    
    list.value = filteredList;
    total.value = filteredList.length; // 注意：这里使用的是过滤后的数量，实际应该从后端获取总数
  } catch (err: any) {
    ElMessage.error(err.message || '查询失败');
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  query.page = 1;
  query.pageSize = 10;
  query.roomId = '';
  query.userId = 0;
  query.fromUserName = '';
  query.toUserName = '';
  query.messageType = undefined;
  loadData();
};

const handlePageChange = (page: number) => {
  query.page = page;
  loadData();
};

const handleSizeChange = (size: number) => {
  query.pageSize = size;
  query.page = 1;
  loadData();
};

const handleDelete = (index: number, row: ChatMessageItem) => {
  ElMessageBox.confirm('确定要删除这条聊天记录吗？删除后无法恢复。', '确认删除', {type: 'warning'})
    .then(async () => {
      try {
        await chatMessageDelete({id: row.id});
        ElMessage.success('删除成功');
        loadData();
      } catch (err: any) {
        ElMessage.error(err.message || '删除失败');
      }
    })
    .catch(() => {});
};

onMounted(loadData);
</script>

<style scoped lang="scss">
.page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mb-12 {
  margin-bottom: 12px;
}

.message-content-image {
  display: inline-block;
}

.message-content-text {
  word-break: break-word;
}

.image-slot {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-placeholder);
}
</style>

