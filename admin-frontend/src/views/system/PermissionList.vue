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
        create-permission="permission:create"
        update-permission="permission:update"
        delete-permission="permission:delete"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
        @onclick-update-row="handleUpdate"
        @onclick-add-row="handleAdd"
      >
        <!-- 自定义操作列 -->
        <template #action="{row}">
          <el-button
            type="primary"
            link
            size="small"
            v-permission="'permission:update'"
            @click="handleAssignMenus(row)"
          >
            {{ t('common.assignMenus') }}
          </el-button>
          <el-button
            type="primary"
            link
            size="small"
            v-permission="'permission:update'"
            @click="handleAssignApis(row)"
          >
            {{ t('common.assignApis') }}
          </el-button>
        </template>
      </D2Table>
    </el-card>

    <!-- 分配菜单对话框 -->
    <el-dialog
      v-model="menuDialogVisible"
      :title="t('common.assignMenus')"
      width="600px"
      @close="handleMenuDialogClose"
    >
      <el-tree
        ref="menuTreeRef"
        :data="menuTreeData"
        :props="{label: 'name', children: 'children'}"
        show-checkbox
        node-key="id"
        :default-checked-keys="selectedMenuIds"
        :check-strictly="false"
      />
      <template #footer>
        <el-button @click="menuDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="menuLoading" @click="handleSaveMenus">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 分配接口对话框 -->
    <el-dialog
      v-model="apiDialogVisible"
      :title="t('common.assignApis')"
      width="600px"
      @close="handleApiDialogClose"
    >
      <el-checkbox-group v-model="selectedApiIds">
        <el-checkbox
          v-for="api in apiListData"
          :key="api.id"
          :label="api.id"
        >
          <el-tag :type="getMethodType(api.method)" size="small" style="margin-right: 8px">
            {{ api.method }}
          </el-tag>
          {{ api.name }} - {{ api.path }}
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="apiDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="apiLoading" @click="handleSaveApis">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {permissionList, permissionCreate, permissionUpdate, permissionDelete, permissionMenuList, permissionMenuUpdate, permissionApiList, permissionApiUpdate, menuTree, apiList} from '@/api/generated/admin';
import type {PermissionItem, MenuItem, ApiItem, PermissionCreateReq, PermissionUpdateReq, PermissionMenuUpdateReq, PermissionApiUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  name: ''
});
const list = ref<PermissionItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 权限-菜单关联相关
const menuDialogVisible = ref(false);
const menuTreeRef = ref();
const menuTreeData = ref<MenuItem[]>([]);
const selectedMenuIds = ref<number[]>([]);
const currentPermissionId = ref<number>(0);
const menuLoading = ref(false);

// 权限-接口关联相关
const apiDialogVisible = ref(false);
const apiListData = ref<ApiItem[]>([]);
const selectedApiIds = ref<number[]>([]);
const apiLoading = ref(false);

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
  {prop: 'code', label: t('common.code')},
  {prop: 'description', label: t('common.description')}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'name', label: t('common.name'), type: D2TableElemType.EditInput, required: true},
  {prop: 'code', label: t('common.code'), type: D2TableElemType.Tag},
  {prop: 'description', label: t('common.description'), type: D2TableElemType.EditInput}
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {prop: 'name', label: t('common.name'), required: true},
  {prop: 'code', label: t('common.code'), required: true},
  {prop: 'description', label: t('common.description')}
]);

const loadData = async () => {
  loading.value = true;
  try {
    const resp = await permissionList({...query});
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

const handleUpdate = async (row: PermissionItem) => {
  try {
    await permissionUpdate(row as PermissionUpdateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    await permissionCreate(row as PermissionCreateReq);
    ElMessage.success('新增成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '新增失败');
  }
};

const handleDelete = (index: number, row: PermissionItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await permissionDelete({id: row.id});
      ElMessage.success(t('common.delete'));
      loadData();
    })
    .catch(() => {});
};

// 加载菜单树
const loadMenus = async () => {
  try {
    const resp = await menuTree();
    menuTreeData.value = resp.list || [];
  } catch (err: any) {
    console.error('Failed to load menus:', err);
  }
};

// 加载接口列表
const loadApis = async () => {
  try {
    const resp = await apiList({page: 1, pageSize: 1000});
    apiListData.value = resp.list || [];
  } catch (err: any) {
    console.error('Failed to load apis:', err);
  }
};

// 打开分配菜单对话框
const handleAssignMenus = async (row: PermissionItem) => {
  currentPermissionId.value = row.id;
  menuDialogVisible.value = true;
  
  // 加载菜单树
  if (menuTreeData.value.length === 0) {
    await loadMenus();
  }
  
  // 加载当前权限的菜单
  try {
    const resp = await permissionMenuList({permissionId: row.id});
    selectedMenuIds.value = resp.menuIds || [];
  } catch (err: any) {
    ElMessage.error(err.message || '加载权限菜单失败');
  }
};

// 保存菜单分配
const handleSaveMenus = async () => {
  menuLoading.value = true;
  try {
    const checkedKeys = menuTreeRef.value?.getCheckedKeys() || [];
    const halfCheckedKeys = menuTreeRef.value?.getHalfCheckedKeys() || [];
    // 合并选中的和半选中的节点
    const allKeys = [...checkedKeys, ...halfCheckedKeys];
    await permissionMenuUpdate({permissionId: currentPermissionId.value, menuIds: allKeys});
    ElMessage.success('菜单分配成功');
    menuDialogVisible.value = false;
  } catch (err: any) {
    ElMessage.error(err.message || '菜单分配失败');
  } finally {
    menuLoading.value = false;
  }
};

// 关闭菜单对话框
const handleMenuDialogClose = () => {
  selectedMenuIds.value = [];
  currentPermissionId.value = 0;
};

// 打开分配接口对话框
const handleAssignApis = async (row: PermissionItem) => {
  currentPermissionId.value = row.id;
  apiDialogVisible.value = true;
  
  // 加载接口列表
  if (apiListData.value.length === 0) {
    await loadApis();
  }
  
  // 加载当前权限的接口
  try {
    const resp = await permissionApiList({permissionId: row.id});
    selectedApiIds.value = resp.apiIds || [];
  } catch (err: any) {
    ElMessage.error(err.message || '加载权限接口失败');
  }
};

// 保存接口分配
const handleSaveApis = async () => {
  apiLoading.value = true;
  try {
    await permissionApiUpdate({permissionId: currentPermissionId.value, apiIds: selectedApiIds.value});
    ElMessage.success('接口分配成功');
    apiDialogVisible.value = false;
  } catch (err: any) {
    ElMessage.error(err.message || '接口分配失败');
  } finally {
    apiLoading.value = false;
  }
};

// 关闭接口对话框
const handleApiDialogClose = () => {
  selectedApiIds.value = [];
  currentPermissionId.value = 0;
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

