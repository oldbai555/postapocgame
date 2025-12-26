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
        create-permission="role:create"
        update-permission="role:update"
        delete-permission="role:delete"
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
        <!-- 自定义操作列 -->
        <template #action="{row}">
          <!-- 超级管理员角色（id=1 或 code='super_admin'）不允许分配权限 -->
          <el-button
            v-if="!isSuperAdminRole(row)"
            v-permission="'role:update'"
            type="primary"
            link
            size="small"
            @click="handleAssignPermissions(row)"
          >
            {{ t('common.assignPermissions') }}
          </el-button>
          <el-tooltip v-else content="超级管理员角色不允许分配权限" placement="top">
            <el-button type="info" link size="small" disabled>
              {{ t('common.assignPermissions') }}
            </el-button>
          </el-tooltip>
        </template>
      </D2Table>

    <!-- 分配权限对话框 -->
    <el-dialog
      v-model="permissionDialogVisible"
      :title="t('common.assignPermissions')"
      width="600px"
      @close="handlePermissionDialogClose"
    >
      <el-tree
        ref="permissionTreeRef"
        :data="permissionTreeData"
        :props="{label: 'name', children: 'children'}"
        show-checkbox
        node-key="id"
        :default-checked-keys="selectedPermissionIds"
        :check-strictly="false"
      />
      <template #footer>
        <el-button @click="permissionDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="permissionLoading" @click="handleSavePermissions">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {roleList, roleCreate, roleUpdate, roleDelete, rolePermissionList, rolePermissionUpdate, permissionList} from '@/api/generated/admin';
import type {RoleItem, PermissionItem, RoleCreateReq, RoleUpdateReq, RolePermissionUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

// 判断是否是超级管理员角色
const isSuperAdminRole = (role: RoleItem): boolean => {
  return role.id === 1 || role.code === 'super_admin';
};

const query = reactive({
  page: 1,
  pageSize: 10,
  name: ''
});
const list = ref<RoleItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 权限分配相关
const permissionDialogVisible = ref(false);
const permissionTreeRef = ref();
const permissionTreeData = ref<any[]>([]);
const selectedPermissionIds = ref<number[]>([]);
const currentRoleId = ref<number>(0);
const permissionLoading = ref(false);

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'name', label: t('common.name')},
  {prop: 'code', label: t('common.code')},
  {prop: 'description', label: t('common.description')},
  {prop: 'status', label: t('common.status'), width: 100}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'name', label: t('common.name'), type: D2TableElemType.EditInput, required: true},
  {prop: 'code', label: t('common.code'), type: D2TableElemType.Tag},
  {prop: 'description', label: t('common.description'), type: D2TableElemType.EditInput},
  {prop: 'status', label: t('common.status'), type: D2TableElemType.Tag}
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {prop: 'name', label: t('common.name'), required: true},
  {prop: 'code', label: t('common.code'), required: true},
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
    const resp = await roleList({...query});
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

const handleUpdate = async (row: RoleItem) => {
  try {
    await roleUpdate(row as RoleUpdateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    await roleCreate(row as RoleCreateReq);
    ElMessage.success('新增成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '新增失败');
  }
};

const handleDelete = (index: number, row: RoleItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await roleDelete({id: row.id});
      ElMessage.success(t('common.delete'));
      loadData();
    })
    .catch(() => {});
};

// 加载权限列表（扁平结构）
const loadPermissions = async () => {
  try {
    const resp = await permissionList({page: 1, pageSize: 1000});
    // 将权限列表转换为树形结构（按模块分组）
    const permissionMap = new Map<string, PermissionItem[]>();
    resp.list.forEach((perm) => {
      const module = perm.code.split(':')[0] || 'other';
      if (!permissionMap.has(module)) {
        permissionMap.set(module, []);
      }
      permissionMap.get(module)!.push(perm);
    });

    const treeData: any[] = [];
    permissionMap.forEach((perms, module) => {
      treeData.push({
        id: `module_${module}`,
        name: getModuleName(module),
        children: perms.map((p) => ({
          id: p.id,
          name: p.name,
          code: p.code
        }))
      });
    });
    permissionTreeData.value = treeData;
  } catch (err: any) {
    ElMessage.error(err.message || '加载权限列表失败');
  }
};

// 获取模块名称
const getModuleName = (module: string): string => {
  const moduleMap: Record<string, string> = {
    role: '角色管理',
    permission: '权限管理',
    department: '部门管理',
    menu: '菜单管理',
    user: '用户管理',
    other: '其他'
  };
  return moduleMap[module] || module;
};

// 打开分配权限对话框
const handleAssignPermissions = async (row: RoleItem) => {
  // 超级管理员角色不允许分配权限
  if (isSuperAdminRole(row)) {
    ElMessage.warning('超级管理员角色不允许分配权限');
    return;
  }
  currentRoleId.value = row.id;
  permissionDialogVisible.value = true;
  
  // 加载权限列表
  if (permissionTreeData.value.length === 0) {
    await loadPermissions();
  }
  
  // 加载当前角色的权限
  try {
    const resp = await rolePermissionList({roleId: row.id});
    selectedPermissionIds.value = resp.permissionIds || [];
  } catch (err: any) {
    ElMessage.error(err.message || '加载角色权限失败');
  }
};

// 保存权限分配
const handleSavePermissions = async () => {
  if (!permissionTreeRef.value) return;
  
  const checkedKeys = permissionTreeRef.value.getCheckedKeys(false) as number[];
  // 过滤掉模块节点
  const permissionIds = checkedKeys.filter((id) => typeof id === 'number');
  
  permissionLoading.value = true;
  try {
    await rolePermissionUpdate({roleId: currentRoleId.value, permissionIds});
    ElMessage.success('权限分配成功');
    permissionDialogVisible.value = false;
  } catch (err: any) {
    ElMessage.error(err.message || '权限分配失败');
  } finally {
    permissionLoading.value = false;
  }
};

// 关闭权限对话框
const handlePermissionDialogClose = () => {
  selectedPermissionIds.value = [];
  currentRoleId.value = 0;
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

