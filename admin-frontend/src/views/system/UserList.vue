<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item :label="t('common.username')">
          <el-input v-model="query.username" :placeholder="t('common.search')" />
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
        create-permission="user:create"
        update-permission="user:update"
        delete-permission="user:delete"
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
          <span v-else-if="column.prop === 'departmentId'">
            {{ getDepartmentName(row.departmentId) }}
          </span>
          <el-avatar
            v-else-if="column.prop === 'avatar'"
            :size="40"
            :src="row.avatar"
          >
            {{ row.username?.charAt(0).toUpperCase() || 'U' }}
          </el-avatar>
        </template>
        <!-- 自定义操作列 -->
        <template #action="{row}">
          <!-- 超级管理员用户（id=1）不允许分配角色 -->
          <el-button
            v-if="!isSuperAdminUser(row)"
            v-permission="'user:update'"
            type="primary"
            link
            size="small"
            @click="handleAssignRoles(row)"
          >
            {{ t('common.assignRoles') }}
          </el-button>
          <el-tooltip v-else content="超级管理员不允许分配角色" placement="top">
            <el-button type="info" link size="small" disabled>
              {{ t('common.assignRoles') }}
            </el-button>
          </el-tooltip>
        </template>
      </D2Table>
    </el-card>

    <!-- 分配角色对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      :title="t('common.assignRoles')"
      width="500px"
      @close="handleRoleDialogClose"
    >
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox
          v-for="role in availableRoles"
          :key="role.id"
          :value="role.id"
        >
          {{ role.name }}
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="roleLoading" @click="handleSaveRoles">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {userList, userCreate, userUpdate, userDelete, userRoleList, userRoleUpdate, roleList as getRoleList, departmentTree} from '@/api/generated/admin';
import type {UserItem, UserCreateReq, UserUpdateReq, DepartmentItem, RoleItem, UserRoleUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

// 判断是否是超级管理员用户
const isSuperAdminUser = (user: UserItem): boolean => {
  return user.id === 1;
};

const query = reactive({
  page: 1,
  pageSize: 10,
  username: ''
});
const list = ref<UserItem[]>([]);
const total = ref(0);
const loading = ref(false);
const departmentList = ref<DepartmentItem[]>([]);

// 角色分配相关
const roleDialogVisible = ref(false);
const roleList = ref<RoleItem[]>([]);
const selectedRoleIds = ref<number[]>([]);
const currentUserId = ref<number>(0);
const roleLoading = ref(false);
const departmentOptions = ref<{label: string; value: number}[]>([]);

// 过滤掉超级管理员角色（id=1），不允许分配
const availableRoles = computed(() => {
  return roleList.value.filter(role => role.id !== 1);
});

// 获取部门名称
const getDepartmentName = (departmentId: number): string => {
  const find = (items: DepartmentItem[], id: number): DepartmentItem | undefined => {
    for (const item of items) {
      if (item.id === id) return item;
      if (item.children) {
        const found = find(item.children, id);
        if (found) return found;
      }
    }
    return undefined;
  };
  return find(departmentList.value, departmentId)?.name || '-';
};

// 扁平化部门树为选项列表
const flattenDepartments = (items: DepartmentItem[], prefix = ''): {label: string; value: number}[] => {
  const result: {label: string; value: number}[] = [];
  items.forEach((item) => {
    result.push({label: prefix + item.name, value: item.id});
    if (item.children && item.children.length > 0) {
      result.push(...flattenDepartments(item.children, prefix + item.name + ' / '));
    }
  });
  return result;
};

// 加载部门列表
const loadDepartments = async () => {
  try {
    const resp = await departmentTree();
    departmentList.value = resp.list || [];
    departmentOptions.value = flattenDepartments(departmentList.value);
  } catch (err: any) {
    console.error('Failed to load departments:', err);
  }
};

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'avatar', label: '头像', width: 100, type: D2TableElemType.Image},
  {prop: 'username', label: t('common.username')},
  {prop: 'signature', label: '个性签名', width: 200},
  {prop: 'departmentId', label: t('common.department')},
  {prop: 'status', label: t('common.status'), width: 100},
  {prop: 'createdAt', label: t('common.createdAt'), width: 180, type: D2TableElemType.ConvertTime}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'username', label: t('common.username'), type: D2TableElemType.EditInput, required: true},
  {prop: 'avatar', label: '头像', type: D2TableElemType.Image},
  {prop: 'signature', label: '个性签名', type: D2TableElemType.EditInput},
  {
    prop: 'departmentId',
    label: t('common.department'),
    type: D2TableElemType.Select,
    options: departmentOptions.value
  },
  {
    prop: 'status',
    label: t('common.status'),
    type: D2TableElemType.Select,
    options: [
      {label: t('status.enabled'), value: 1},
      {label: t('status.disabled'), value: 0}
    ]
  },
  {prop: 'createdAt', label: t('common.createdAt'), type: D2TableElemType.ConvertTime}
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {prop: 'username', label: t('common.username'), required: true},
  {prop: 'password', label: t('common.password'), required: true},
  {prop: 'avatar', label: '头像', type: D2TableElemType.Image},
  {prop: 'signature', label: '个性签名'},
  {
    prop: 'departmentId',
    label: t('common.department'),
    type: D2TableElemType.Select,
    options: departmentOptions.value
  },
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
    const resp = await userList({...query});
    list.value = resp.list;
    total.value = resp.total;
  } catch (err: any) {
    ElMessage.error(err.message || t('common.loadFailed'));
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  query.page = 1;
  query.pageSize = 10;
  query.username = '';
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

const handleUpdate = async (row: UserItem) => {
  try {
    // 如果密码为空，则不传递密码字段
    const updateData: any = {...row};
    if (!updateData.password || updateData.password.trim() === '') {
      delete updateData.password;
    }
    await userUpdate(updateData as UserUpdateReq);
    ElMessage.success(t('common.updateSuccess'));
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || t('common.submitFailed'));
  }
};

const handleAdd = async (row: any) => {
  try {
    if (!row.password || row.password.trim() === '') {
      ElMessage.error(t('common.passwordRequired'));
      return;
    }
    await userCreate(row as UserCreateReq);
    ElMessage.success(t('common.createSuccess'));
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || t('common.submitFailed'));
  }
};

const handleDelete = (index: number, row: UserItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await userDelete({id: row.id});
      ElMessage.success(t('common.deleteSuccess'));
      loadData();
    })
    .catch(() => {});
};

// 加载角色列表
const loadRoles = async () => {
  try {
    const resp = await getRoleList({page: 1, pageSize: 1000});
    roleList.value = resp.list || [];
    if (roleList.value.length === 0) {
      ElMessage.warning('暂无可用角色，请先在角色管理中创建角色');
    }
  } catch (err: any) {
    console.error('Failed to load roles:', err);
    throw err; // 重新抛出错误，让调用者处理
  }
};

// 打开分配角色对话框
const handleAssignRoles = async (row: UserItem) => {
  // 超级管理员用户不允许分配角色
  if (isSuperAdminUser(row)) {
    ElMessage.warning('超级管理员不允许分配角色');
    return;
  }
  currentUserId.value = row.id;
  roleDialogVisible.value = true;
  
  // 每次打开对话框时都重新加载角色列表，确保数据最新
  try {
    await loadRoles();
  } catch (err: any) {
    ElMessage.error(err.message || '加载角色列表失败');
    return;
  }
  
  // 加载当前用户的角色
  try {
    const resp = await userRoleList({userId: row.id});
    selectedRoleIds.value = resp.roleIds || [];
  } catch (err: any) {
    ElMessage.error(err.message || '加载用户角色失败');
  }
};

// 保存角色分配
const handleSaveRoles = async () => {
  // 检查是否包含超级管理员角色
  if (selectedRoleIds.value.includes(1)) {
    ElMessage.warning('不允许分配超级管理员角色');
    return;
  }
  
  roleLoading.value = true;
  try {
    await userRoleUpdate({userId: currentUserId.value, roleIds: selectedRoleIds.value});
    ElMessage.success('角色分配成功');
    roleDialogVisible.value = false;
  } catch (err: any) {
    ElMessage.error(err.message || '角色分配失败');
  } finally {
    roleLoading.value = false;
  }
};

// 关闭角色对话框
const handleRoleDialogClose = () => {
  selectedRoleIds.value = [];
  currentUserId.value = 0;
};

onMounted(async () => {
  await loadDepartments();
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

