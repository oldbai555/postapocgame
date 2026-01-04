<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="群组名称">
          <el-input v-model="query.name" placeholder="搜索群组名称" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">搜索</el-button>
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
        :drawer-add-columns="drawerAddColumns"
        :have-edit="true"
        :have-detail="true"
        create-permission="chat:group:create"
        update-permission="chat:group:update"
        delete-permission="chat:group:delete"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
        @onclick-update-row="handleUpdate"
        @onclick-add-row="handleAdd"
      >
        <!-- 自定义头像列 -->
        <template #cell="{row, column}">
          <el-avatar
            v-if="column.prop === 'avatar'"
            :size="40"
            :src="row.avatar"
          >
            {{ row.name?.charAt(0) || 'G' }}
          </el-avatar>
        </template>
        <!-- 自定义操作列 -->
        <template #action="{row}">
          <el-button
            v-permission="'chat:group:member'"
            type="primary"
            link
            size="small"
            @click="handleManageMembers(row)"
          >
            成员管理
          </el-button>
        </template>
      </D2Table>
    </el-card>

    <!-- 成员管理对话框 -->
    <el-dialog
      v-model="memberDialogVisible"
      :title="`成员管理 - ${currentGroupName}`"
      width="800px"
      @close="handleMemberDialogClose"
    >
      <div class="member-dialog-content">
        <!-- 成员列表 -->
        <div class="member-list-section">
          <div class="section-header">
            <span>当前成员 ({{ memberList.length }})</span>
            <el-button
              v-permission="'chat:group:member'"
              type="primary"
              size="small"
              @click="showAddMemberDialog = true"
            >
              添加成员
            </el-button>
          </div>
          <el-table :data="memberList" style="width: 100%">
            <el-table-column prop="avatar" label="头像" width="80">
              <template #default="{row}">
                <el-avatar :size="40" :src="row.avatar">
                  {{ row.username?.charAt(0).toUpperCase() || 'U' }}
                </el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="username" label="用户名" />
            <el-table-column prop="nickname" label="昵称" />
            <el-table-column prop="departmentName" label="部门" />
            <el-table-column prop="roleNames" label="角色">
              <template #default="{row}">
                <el-tag
                  v-for="role in row.roleNames"
                  :key="role"
                  size="small"
                  style="margin-right: 4px"
                >
                  {{ role }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="joinedAt" label="加入时间" width="180">
              <template #default="{row}">
                {{ formatTime(row.joinedAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{row}">
                <el-button
                  v-permission="'chat:group:member'"
                  type="danger"
                  link
                  size="small"
                  @click="handleRemoveMember(row)"
                >
                  移除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
      <template #footer>
        <el-button @click="memberDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 添加成员对话框 -->
    <el-dialog
      v-model="showAddMemberDialog"
      title="添加成员"
      width="500px"
      @close="handleAddMemberDialogClose"
    >
      <el-select
        v-model="selectedUserIds"
        multiple
        filterable
        placeholder="选择要添加的用户"
        style="width: 100%"
      >
        <el-option
          v-for="user in availableUsers"
          :key="user.id"
          :label="`${user.departmentName || ''} - ${user.roleNames?.join(',') || ''} - ${user.nickname || user.username}`"
          :value="user.id"
        />
      </el-select>
      <template #footer>
        <el-button @click="showAddMemberDialog = false">取消</el-button>
        <el-button type="primary" :loading="memberLoading" @click="handleAddMembers">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {
  chatGroupList,
  chatGroupCreate,
  chatGroupUpdate,
  chatGroupDelete,
  chatGroupDetail,
  chatGroupMemberList,
  chatGroupMemberAdd,
  chatGroupMemberRemove,
  userList
} from '@/api/generated/admin';
import type {
  ChatGroupItem,
  ChatGroupCreateReq,
  ChatGroupUpdateReq,
  ChatGroupDetailResp,
  ChatGroupMemberItem
} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  name: ''
});
const list = ref<ChatGroupItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 成员管理相关
const memberDialogVisible = ref(false);
const showAddMemberDialog = ref(false);
const currentGroupId = ref<number>(0);
const currentGroupName = ref<string>('');
const memberList = ref<ChatGroupMemberItem[]>([]);
const availableUsers = ref<any[]>([]);
const selectedUserIds = ref<number[]>([]);
const memberLoading = ref(false);

// 格式化时间
const formatTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  const date = new Date(timestamp * 1000);
  return date.toLocaleString('zh-CN');
};

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'avatar', label: '头像', width: 100, type: D2TableElemType.Image},
  {prop: 'name', label: '群组名称'},
  {prop: 'description', label: '描述', width: 200},
  {prop: 'createdAt', label: '创建时间', width: 180, type: D2TableElemType.ConvertTime}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'name', label: '群组名称', type: D2TableElemType.EditInput, required: true},
  {prop: 'avatar', label: '头像', type: D2TableElemType.Image},
  {prop: 'description', label: '描述', type: D2TableElemType.EditInput},
  {prop: 'createdAt', label: '创建时间', type: D2TableElemType.ConvertTime}
]);

// 新增抽屉列配置
const drawerAddColumns = computed<DrawerColumn[]>(() => [
  {prop: 'name', label: '群组名称', required: true},
  {prop: 'avatar', label: '头像', type: D2TableElemType.Image},
  {prop: 'description', label: '描述'}
]);

// 加载群组列表
const loadData = async () => {
  loading.value = true;
  try {
    const resp = await chatGroupList({
      page: query.page,
      pageSize: query.pageSize,
      name: query.name || undefined
    });
    list.value = resp.list;
    total.value = resp.total;
  } catch (err: any) {
    ElMessage.error(err.message || '加载失败');
  } finally {
    loading.value = false;
  }
};

// 加载用户列表（用于添加成员）
const loadUsers = async () => {
  try {
    const resp = await userList({page: 1, pageSize: 1000});
    // 过滤掉已经在群组中的用户
    const memberUserIds = memberList.value.map(m => m.userId);
    availableUsers.value = resp.list
      .filter(user => !memberUserIds.includes(user.id))
      .map(user => ({
        id: user.id,
        username: user.username,
        nickname: user.nickname,
        departmentName: user.departmentName || '',
        roleNames: user.roleNames || []
      }));
  } catch (err: any) {
    ElMessage.error(err.message || '加载用户列表失败');
  }
};

// 加载群组成员列表
const loadMembers = async (groupId: number) => {
  memberLoading.value = true;
  try {
    const resp = await chatGroupMemberList({}, groupId);
    memberList.value = resp.list;
    // 重新加载用户列表（排除已加入的成员）
    await loadUsers();
  } catch (err: any) {
    ElMessage.error(err.message || '加载成员列表失败');
  } finally {
    memberLoading.value = false;
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

const handleUpdate = async (row: ChatGroupItem) => {
  try {
    await chatGroupUpdate({
      id: row.id,
      name: row.name,
      avatar: row.avatar,
      description: row.description
    } as ChatGroupUpdateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    await chatGroupCreate({
      name: row.name,
      avatar: row.avatar || '',
      description: row.description || '',
      userIds: [] // 初始成员为空，创建后可以添加
    } as ChatGroupCreateReq);
    ElMessage.success('创建成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '创建失败');
  }
};

const handleDelete = (index: number, row: ChatGroupItem) => {
  ElMessageBox.confirm('确定要删除这个群组吗？', '确认删除', {type: 'warning'})
    .then(async () => {
      await chatGroupDelete({id: row.id});
      ElMessage.success('删除成功');
      loadData();
    })
    .catch(() => {});
};

// 成员管理
const handleManageMembers = async (row: ChatGroupItem) => {
  currentGroupId.value = row.id;
  currentGroupName.value = row.name || '';
  memberDialogVisible.value = true;
  await loadMembers(row.id);
};

const handleMemberDialogClose = () => {
  memberList.value = [];
  currentGroupId.value = 0;
  currentGroupName.value = '';
};

const handleAddMemberDialogClose = () => {
  selectedUserIds.value = [];
};

const handleAddMembers = async () => {
  if (selectedUserIds.value.length === 0) {
    ElMessage.warning('请选择要添加的用户');
    return;
  }
  memberLoading.value = true;
  try {
    await chatGroupMemberAdd({
      chatId: currentGroupId.value,
      userIds: selectedUserIds.value
    });
    ElMessage.success('添加成员成功');
    showAddMemberDialog.value = false;
    selectedUserIds.value = [];
    await loadMembers(currentGroupId.value);
  } catch (err: any) {
    ElMessage.error(err.message || '添加成员失败');
  } finally {
    memberLoading.value = false;
  }
};

const handleRemoveMember = (member: ChatGroupMemberItem) => {
  ElMessageBox.confirm(`确定要将 ${member.nickname || member.username} 移出群组吗？`, '确认移除', {type: 'warning'})
    .then(async () => {
      try {
        await chatGroupMemberRemove({
          chatId: currentGroupId.value,
          userId: member.userId
        });
        ElMessage.success('移除成员成功');
        await loadMembers(currentGroupId.value);
      } catch (err: any) {
        ElMessage.error(err.message || '移除成员失败');
      }
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
.member-dialog-content {
  max-height: 500px;
  overflow-y: auto;
}
.member-list-section {
  margin-bottom: 20px;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 500;
}
</style>

