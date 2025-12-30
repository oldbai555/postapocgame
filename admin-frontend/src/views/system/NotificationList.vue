<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="消息来源">
          <el-select v-model="query.sourceType" placeholder="请选择消息来源" clearable style="width: 150px">
            <el-option
              v-for="item in sourceTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="已读状态">
          <el-select v-model="query.readStatus" placeholder="请选择已读状态" clearable style="width: 120px">
            <el-option label="未读" :value="0" />
            <el-option label="已读" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">{{ t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 操作按钮 -->
    <el-card class="mb-12">
      <el-space>
        <el-button type="success" :loading="readAllLoading" @click="handleReadAll">
          全部已读
        </el-button>
        <el-button type="warning" :loading="clearReadLoading" @click="handleClearRead">
          清除已读
        </el-button>
      </el-space>
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
        :have-edit="false"
        :have-detail="true"
        :have-add="false"
        delete-permission=""
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
      >
        <!-- 自定义列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'sourceType'" :type="getSourceTypeTag(row.sourceType)">
            {{ getSourceTypeLabel(row.sourceType) }}
          </el-tag>
          <el-tag v-else-if="column.prop === 'readStatus'" :type="row.readStatus === 1 ? 'success' : 'warning'">
            {{ row.readStatus === 1 ? '已读' : '未读' }}
          </el-tag>
          <span v-else-if="column.prop === 'readAt'">
            {{ row.readAt ? formatTime(row.readAt) : '-' }}
          </span>
          <el-tooltip v-else-if="column.prop === 'content'" :content="row.content" placement="top" :disabled="!row.content || row.content.length <= 50">
            <div class="content-cell">
              {{ row.content }}
            </div>
          </el-tooltip>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import { notificationList, notificationDelete, notificationReadAll, notificationClearRead, dictGet } from '@/api/generated/admin';
import type { NotificationItem, NotificationDeleteReq } from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  sourceType: '',
  readStatus: 0
});
const list = ref<NotificationItem[]>([]);
const total = ref(0);
const loading = ref(false);
const readAllLoading = ref(false);
const clearReadLoading = ref(false);
const sourceTypeOptions = ref<Array<{label: string; value: string}>>([]);

// 加载消息来源字典
const loadSourceTypeOptions = async () => {
  try {
    const resp = await dictGet({code: 'notification_source_type'});
    if (resp && resp.items) {
      sourceTypeOptions.value = resp.items.map((item: any) => ({
        label: item.label,
        value: item.value
      }));
    }
  } catch (err: any) {
    console.error('加载消息来源字典失败:', err);
    // 如果字典加载失败，使用默认值
    sourceTypeOptions.value = [
      {label: '在线聊天', value: 'chat'},
      {label: '系统公告', value: 'notice'},
      {label: '系统通知', value: 'system'}
    ];
  }
};

// 获取消息来源标签
const getSourceTypeLabel = (sourceType: string): string => {
  const option = sourceTypeOptions.value.find(item => item.value === sourceType);
  return option ? option.label : sourceType;
};

// 获取消息来源标签颜色
const getSourceTypeTag = (sourceType: string): string | undefined => {
  const map: Record<string, string> = {
    'chat': 'primary',
    'notice': 'warning',
    'system': 'info'
  };
  return map[sourceType] || undefined;
};

// 格式化时间
const formatTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  const date = new Date(timestamp * 1000);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'sourceType', label: '消息来源', width: 120},
  {prop: 'title', label: '消息标题', minWidth: 200},
  {prop: 'content', label: '消息内容', minWidth: 300},
  {prop: 'readStatus', label: '已读状态', width: 100},
  {prop: 'readAt', label: '已读时间', width: 180},
  {prop: 'createdAt', label: '创建时间', width: 180, type: D2TableElemType.ConvertTime}
]);

// 详情抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'sourceType', label: '消息来源', type: D2TableElemType.Tag},
  {prop: 'title', label: '消息标题'},
  {prop: 'content', label: '消息内容'},
  {prop: 'readStatus', label: '已读状态', type: D2TableElemType.Tag},
  {prop: 'readAt', label: '已读时间'},
  {prop: 'createdAt', label: '创建时间', type: D2TableElemType.ConvertTime}
]);

const loadData = async () => {
  loading.value = true;
  try {
    const req: any = {
      page: query.page,
      pageSize: query.pageSize
    };
    if (query.sourceType) {
      req.sourceType = query.sourceType;
    }
    if (query.readStatus >= 0) {
      req.readStatus = query.readStatus;
    }
    const resp = await notificationList(req);
    list.value = resp.list;
    total.value = resp.total;
  } catch (err: any) {
    ElMessage.error(err.message || t('common.searchFailed'));
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  query.page = 1;
  query.pageSize = 10;
  query.sourceType = '';
  query.readStatus = 0;
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

const handleDelete = (index: number, row: NotificationItem) => {
  ElMessageBox.confirm('确定要删除该消息通知吗？', '确认删除', {type: 'warning'})
    .then(async () => {
      await notificationDelete({id: row.id});
      ElMessage.success('删除成功');
      loadData();
    })
    .catch(() => {});
};

const handleReadAll = async () => {
  readAllLoading.value = true;
  try {
    await notificationReadAll();
    ElMessage.success('全部已读操作成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败');
  } finally {
    readAllLoading.value = false;
  }
};

const handleClearRead = async () => {
  ElMessageBox.confirm('确定要清除所有已读消息吗？此操作不可恢复。', '确认清除', {type: 'warning'})
    .then(async () => {
      clearReadLoading.value = true;
      try {
        await notificationClearRead();
        ElMessage.success('清除已读消息成功');
        loadData();
      } catch (err: any) {
        ElMessage.error(err.message || '操作失败');
      } finally {
        clearReadLoading.value = false;
      }
    })
    .catch(() => {});
};

onMounted(async () => {
  await loadSourceTypeOptions();
  loadData();
});
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.mb-12 {
  margin-bottom: 12px;
}
.content-cell {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: default;
}
</style>

