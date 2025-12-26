<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="字典类型">
          <el-select v-model="query.typeId" placeholder="请选择字典类型" clearable style="width: 200px">
            <el-option
              v-for="item in dictTypeOptions"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="字典标签">
          <el-input v-model="query.label" placeholder="请输入字典标签" />
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
        create-permission="dict_item:create"
        update-permission="dict_item:update"
        delete-permission="dict_item:delete"
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
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {dictItemList, dictItemCreate, dictItemUpdate, dictItemDelete, dictTypeList} from '@/api/generated/admin';
import type {DictItemItem, DictItemCreateReq, DictItemUpdateReq, DictTypeItem} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  typeId: undefined as number | undefined,
  label: ''
});
const list = ref<DictItemItem[]>([]);
const total = ref(0);
const loading = ref(false);
const dictTypeOptions = ref<DictTypeItem[]>([]);

// 加载字典类型选项
const loadDictTypes = async () => {
  try {
    const resp = await dictTypeList({page: 1, pageSize: 1000});
    dictTypeOptions.value = resp.list;
  } catch (err: any) {
    console.error('加载字典类型失败:', err);
  }
};

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'typeId', label: '字典类型', width: 120},
  {prop: 'label', label: '字典标签'},
  {prop: 'value', label: '字典值'},
  {prop: 'sort', label: '排序', width: 80},
  {prop: 'status', label: t('common.status'), width: 100},
  {prop: 'remark', label: '备注'}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'typeId', label: '字典类型', type: D2TableElemType.Tag},
  {prop: 'label', label: '字典标签', type: D2TableElemType.EditInput, required: true},
  {prop: 'value', label: '字典值', type: D2TableElemType.EditInput, required: true},
  {prop: 'sort', label: '排序', type: D2TableElemType.EditInput},
  {
    prop: 'status',
    label: t('common.status'),
    type: D2TableElemType.Select,
    options: [
      {label: t('status.enabled'), value: 1},
      {label: t('status.disabled'), value: 0}
    ]
  },
  {prop: 'remark', label: '备注', type: D2TableElemType.EditInput}
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {
    prop: 'typeId',
    label: '字典类型',
    type: D2TableElemType.Select,
    required: true,
    options: dictTypeOptions.value.map(item => ({label: item.name, value: item.id}))
  },
  {prop: 'label', label: '字典标签', required: true},
  {prop: 'value', label: '字典值', required: true},
  {prop: 'sort', label: '排序'},
  {
    prop: 'status',
    label: t('common.status'),
    type: D2TableElemType.Select,
    options: [
      {label: t('status.enabled'), value: 1},
      {label: t('status.disabled'), value: 0}
    ]
  },
  {prop: 'remark', label: '备注'}
]);

const loadData = async () => {
  loading.value = true;
  try {
    const resp = await dictItemList({...query});
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
  query.typeId = undefined;
  query.label = '';
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

const handleUpdate = async (row: DictItemItem) => {
  try {
    await dictItemUpdate(row as DictItemUpdateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    await dictItemCreate(row as DictItemCreateReq);
    ElMessage.success('新增成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '新增失败');
  }
};

const handleDelete = (index: number, row: DictItemItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await dictItemDelete({id: row.id});
      ElMessage.success(t('common.delete'));
      loadData();
    })
    .catch(() => {});
};

onMounted(() => {
  loadDictTypes();
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
</style>

