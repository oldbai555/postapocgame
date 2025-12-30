<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="公告标题">
          <el-input v-model="query.title" placeholder="请输入公告标题" clearable />
        </el-form-item>
        <el-form-item label="公告类型">
          <el-select v-model="query.noticeType" placeholder="请选择公告类型" clearable style="width: 150px">
            <el-option label="普通公告" :value="1" />
            <el-option label="重要公告" :value="2" />
            <el-option label="紧急公告" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="请选择状态" clearable style="width: 120px">
            <el-option label="草稿" :value="1" />
            <el-option label="已发布" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">{{ t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ t('common.reset') }}</el-button>
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
        :drawer-add-columns="drawerAddColumns"
        :have-edit="true"
        :have-detail="true"
        create-permission="notice:create"
        update-permission="notice:update"
        delete-permission="notice:delete"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
        @onclick-update-row="handleUpdate"
        @onclick-add-row="handleAdd"
      >
        <!-- 自定义列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'type'" :type="getNoticeTypeTag(row.type) || undefined">
            {{ getNoticeTypeLabel(row.type) }}
          </el-tag>
          <el-tag v-else-if="column.prop === 'status'" :type="row.status === 2 ? 'success' : 'info'">
            {{ row.status === 2 ? '已发布' : '草稿' }}
          </el-tag>
          <span v-else-if="column.prop === 'publishTime'">
            {{ row.publishTime ? formatTime(row.publishTime) : '-' }}
          </span>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import { noticeList, noticeCreate, noticeUpdate, noticeDelete } from '@/api/generated/admin';
import type { NoticeItem, NoticeCreateReq, NoticeUpdateReq, NoticeDeleteReq } from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  title: '',
  noticeType: undefined as number | undefined,
  status: undefined as number | undefined
});
const list = ref<NoticeItem[]>([]);
const total = ref(0);
const loading = ref(false);
const originalRowMap = ref<Map<number, NoticeItem>>(new Map()); // 保存原始行数据

// 获取公告类型标签
const getNoticeTypeLabel = (type: number): string => {
  const map: Record<number, string> = {
    1: '普通公告',
    2: '重要公告',
    3: '紧急公告'
  };
  return map[type] || '未知';
};

// 获取公告类型标签颜色
const getNoticeTypeTag = (type: number): string | undefined => {
  const map: Record<number, string> = {
    1: 'info',
    2: 'warning',
    3: 'danger'
  };
  return map[type] || undefined;
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
  {prop: 'title', label: '公告标题', minWidth: 200},
  {prop: 'type', label: '公告类型', width: 120},
  {prop: 'status', label: '状态', width: 100},
  {prop: 'publishTime', label: '发布时间', width: 180},
  {prop: 'createdAt', label: t('common.createdAt'), width: 180}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'title', label: '公告标题', type: D2TableElemType.EditInput, required: true},
  {prop: 'content', label: '公告内容', type: D2TableElemType.EditTextarea, required: true},
  {
    prop: 'type',
    label: '公告类型',
    type: D2TableElemType.Select,
    options: [
      {label: '普通公告', value: 1},
      {label: '重要公告', value: 2},
      {label: '紧急公告', value: 3}
    ]
  },
  {
    prop: 'status',
    label: '状态',
    type: D2TableElemType.Select,
    options: [
      {label: '草稿', value: 1},
      {label: '已发布', value: 2}
    ]
  },
  {prop: 'publishTime', label: '发布时间', type: D2TableElemType.DateTime},
  {prop: 'createdAt', label: t('common.createdAt')}
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {prop: 'title', label: '公告标题', required: true},
  {prop: 'content', label: '公告内容', type: D2TableElemType.EditTextarea, required: true},
  {
    prop: 'type',
    label: '公告类型',
    type: D2TableElemType.Select,
    options: [
      {label: '普通公告', value: 1},
      {label: '重要公告', value: 2},
      {label: '紧急公告', value: 3}
    ]
  },
  {
    prop: 'status',
    label: '状态',
    type: D2TableElemType.Select,
    options: [
      {label: '草稿', value: 1},
      {label: '已发布', value: 2}
    ]
  },
  {prop: 'publishTime', label: '发布时间（留空则立即发布）', type: D2TableElemType.DateTime}
]);

const loadData = async () => {
  loading.value = true;
  try {
    const req: any = {
      page: query.page,
      pageSize: query.pageSize,
      title: query.title || undefined
    };
    if (query.noticeType !== undefined && query.noticeType > 0) {
      req.type = query.noticeType;
    }
    if (query.status !== undefined && query.status >= 0) {
      req.status = query.status;
    }
    const resp = await noticeList(req);
    list.value = resp.list;
    total.value = resp.total;
    // 保存原始行数据用于更新时比较
    originalRowMap.value.clear();
    resp.list.forEach((item: NoticeItem) => {
      originalRowMap.value.set(item.id, {...item});
    });
  } catch (err: any) {
    ElMessage.error(err.message || t('common.searchFailed'));
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  query.page = 1;
  query.pageSize = 10;
  query.title = '';
  query.noticeType = undefined;
  query.status = undefined;
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

const handleUpdate = async (row: NoticeItem) => {
  try {
    // 获取原始行数据
    const originalRow = originalRowMap.value.get(row.id);
    
    // 如果原始状态是已发布，不允许修改状态
    if (originalRow && originalRow.status === 2 && row.status !== 2) {
      ElMessage.warning('已发布的公告不允许修改状态，只能删除');
      // 恢复原始状态
      row.status = originalRow.status;
      return;
    }
    
    // 如果从草稿变为发布，需要二次确认
    if (originalRow && originalRow.status === 1 && row.status === 2) {
      try {
        await ElMessageBox.confirm(
          '确定要发布该公告吗？发布后所有用户都能看到，且无法再修改状态。',
          '确认发布',
          {
            type: 'warning',
            confirmButtonText: '确定发布',
            cancelButtonText: '取消'
          }
        );
      } catch {
        // 用户取消发布，恢复原始状态
        row.status = originalRow.status;
        return;
      }
    }
    
    // 如果原始状态是已发布，不允许修改任何字段（除了删除）
    if (originalRow && originalRow.status === 2) {
      // 只允许修改标题和内容，不允许修改状态、类型和发布时间
      const updateReq: NoticeUpdateReq = {
        id: row.id,
        title: row.title,
        content: row.content,
        type: originalRow.type, // 保持原始类型
        status: originalRow.status, // 保持原始状态
        publishTime: originalRow.publishTime // 保持原始发布时间
      };
      await noticeUpdate(updateReq);
      ElMessage.success('更新成功');
      loadData();
      return;
    }
    
    const updateReq: NoticeUpdateReq = {
      id: row.id,
      title: row.title,
      content: row.content,
      type: row.type,
      status: row.status,
      publishTime: row.publishTime
    };
    await noticeUpdate(updateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    const createReq: NoticeCreateReq = {
      title: row.title,
      content: row.content,
      type: row.type || 1,
      status: row.status !== undefined ? row.status : 1, // 默认草稿
      publishTime: row.publishTime || 0
    };
    await noticeCreate(createReq);
    ElMessage.success('新增成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '新增失败');
  }
};

const handleDelete = (index: number, row: NoticeItem) => {
  ElMessageBox.confirm('确定要删除该公告吗？', '确认删除', {type: 'warning'})
    .then(async () => {
      await noticeDelete({id: row.id});
      ElMessage.success('删除成功');
      loadData();
    })
    .catch(() => {});
};

onMounted(loadData);
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
</style>

