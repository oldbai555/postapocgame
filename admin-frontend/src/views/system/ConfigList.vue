<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="配置分组">
          <el-input v-model="query.group" placeholder="请输入配置分组" />
        </el-form-item>
        <el-form-item label="配置键">
          <el-input v-model="query.key" placeholder="请输入配置键" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">{{ t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ t('common.reset') }}</el-button>
          <el-button
            type="warning"
            :loading="refreshLoading"
            @click="handleRefreshCache"
            icon="Refresh"
            v-permission="'config:update'"
          >
            刷新缓存
          </el-button>
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
        create-permission="config:create"
        update-permission="config:update"
        delete-permission="config:delete"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
        @onclick-update-row="handleUpdate"
        @onclick-add-row="handleAdd"
      >
        <!-- 自定义配置类型列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'type'" :type="getConfigTypeTag(row.type)">
            {{ row.type }}
          </el-tag>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {configList, configCreate, configUpdate, configDelete, cacheRefresh} from '@/api/generated/admin';
import type {ConfigItem, ConfigCreateReq, ConfigUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  group: '',
  key: ''
});
const list = ref<ConfigItem[]>([]);
const total = ref(0);
const loading = ref(false);
const refreshLoading = ref(false);

// 获取配置类型的标签类型
const getConfigTypeTag = (type: string): string => {
  const typeMap: Record<string, string> = {
    'string': 'success',
    'number': 'warning',
    'boolean': 'primary',
    'json': 'info'
  };
  return typeMap[type] || 'info';
};

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'group', label: '配置分组', width: 120},
  {prop: 'key', label: '配置键'},
  {prop: 'value', label: '配置值'},
  {prop: 'type', label: '配置类型', width: 100},
  {prop: 'description', label: t('common.description')}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'group', label: '配置分组', type: D2TableElemType.Tag},
  {prop: 'key', label: '配置键', type: D2TableElemType.Tag},
  {prop: 'value', label: '配置值', type: D2TableElemType.EditInput, required: true},
  {prop: 'type', label: '配置类型', type: D2TableElemType.Tag},
  {prop: 'description', label: t('common.description'), type: D2TableElemType.EditInput}
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {prop: 'group', label: '配置分组', required: true},
  {prop: 'key', label: '配置键', required: true},
  {prop: 'value', label: '配置值', required: true},
  {
    prop: 'type',
    label: '配置类型',
    type: D2TableElemType.Select,
    options: [
      {label: 'string', value: 'string'},
      {label: 'number', value: 'number'},
      {label: 'boolean', value: 'boolean'},
      {label: 'json', value: 'json'}
    ]
  },
  {prop: 'description', label: t('common.description')}
]);

const loadData = async () => {
  loading.value = true;
  try {
    const resp = await configList({...query});
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
  query.group = '';
  query.key = '';
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

const handleUpdate = async (row: ConfigItem) => {
  try {
    await configUpdate(row as ConfigUpdateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    await configCreate(row as ConfigCreateReq);
    ElMessage.success('新增成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '新增失败');
  }
};

const handleDelete = (index: number, row: ConfigItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await configDelete({id: row.id});
      ElMessage.success(t('common.delete'));
      loadData();
    })
    .catch(() => {});
};

// 刷新缓存
const handleRefreshCache = async () => {
  try {
    refreshLoading.value = true;
    await cacheRefresh();
    ElMessage.success('缓存刷新成功');
  } catch (error: any) {
    ElMessage.error(error.message || '缓存刷新失败');
  } finally {
    refreshLoading.value = false;
  }
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

