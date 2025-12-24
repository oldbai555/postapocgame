<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item :label="t('common.name')">
          <el-input v-model="query.name" :placeholder="t('common.search')" />
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
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
        @onclick-update-row="handleUpdate"
        @onclick-add-row="handleAdd"
      >
        <!-- 自定义状态列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'status'" :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? t('status.enabled') : t('status.disabled') }}
          </el-tag>
          <el-tag v-else-if="column.prop === 'method'" :type="getMethodType(row.method)">
            {{ row.method }}
          </el-tag>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {apiList, apiCreate, apiUpdate, apiDelete} from '@/api/generated/admin';
import type {ApiItem, ApiCreateReq, ApiUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  name: ''
});
const list = ref<ApiItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 获取HTTP方法的标签类型
const getMethodType = (method: string): string => {
  const methodMap: Record<string, string> = {
    'GET': 'success',
    'POST': 'warning',
    'PUT': 'primary',
    'DELETE': 'danger',
    'PATCH': 'info'
  };
  return methodMap[method] || 'info';
};

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'name', label: t('common.name')},
  {prop: 'method', label: 'HTTP方法', width: 120},
  {prop: 'path', label: '接口路径'},
  {prop: 'description', label: t('common.description')},
  {prop: 'status', label: t('common.status'), width: 100}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'name', label: t('common.name'), type: D2TableElemType.EditInput, required: true},
  {
    prop: 'method',
    label: 'HTTP方法',
    type: D2TableElemType.Select,
    required: true,
    options: [
      {label: 'GET', value: 'GET'},
      {label: 'POST', value: 'POST'},
      {label: 'PUT', value: 'PUT'},
      {label: 'DELETE', value: 'DELETE'},
      {label: 'PATCH', value: 'PATCH'}
    ]
  },
  {prop: 'path', label: '接口路径', type: D2TableElemType.EditInput, required: true},
  {prop: 'description', label: t('common.description'), type: D2TableElemType.EditInput},
  {
    prop: 'status',
    label: t('common.status'),
    type: D2TableElemType.Select,
    options: [
      {label: t('status.enabled'), value: 1},
      {label: t('status.disabled'), value: 0}
    ]
  }
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {prop: 'name', label: t('common.name'), required: true},
  {
    prop: 'method',
    label: 'HTTP方法',
    type: D2TableElemType.Select,
    required: true,
    options: [
      {label: 'GET', value: 'GET'},
      {label: 'POST', value: 'POST'},
      {label: 'PUT', value: 'PUT'},
      {label: 'DELETE', value: 'DELETE'},
      {label: 'PATCH', value: 'PATCH'}
    ]
  },
  {prop: 'path', label: '接口路径', required: true},
  {prop: 'description', label: t('common.description')},
  {
    prop: 'status',
    label: t('common.status'),
    type: D2TableElemType.Select,
    options: [
      {label: t('status.enabled'), value: 1},
      {label: t('status.disabled'), value: 0}
    ]
  }
]);

const loadData = async () => {
  loading.value = true;
  try {
    const resp = await apiList({...query});
    list.value = resp.list;
    total.value = resp.total;
  } catch (err: any) {
    ElMessage.error(err.message || t('common.search'));
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  query.page = 1;
  query.pageSize = 10;
  query.name = '';
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

const handleUpdate = async (row: ApiItem) => {
  try {
    await apiUpdate(row as ApiUpdateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    await apiCreate(row as ApiCreateReq);
    ElMessage.success('新增成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '新增失败');
  }
};

const handleDelete = (index: number, row: ApiItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await apiDelete({id: row.id});
      ElMessage.success(t('common.delete'));
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

