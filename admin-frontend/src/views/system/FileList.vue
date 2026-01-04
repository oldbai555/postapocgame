<template>
  <div class="page">
    <!-- 搜索表单和上传按钮 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item :label="t('common.name')">
          <el-input v-model="query.name" :placeholder="t('common.search')" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">{{ t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ t('common.reset') }}</el-button>
          <el-upload
            :action="uploadUrl"
            :headers="uploadHeaders"
            :on-success="handleUploadSuccess"
            :on-error="handleUploadError"
            :before-upload="beforeUpload"
            :show-file-list="false"
            style="display: inline-block; margin-left: 10px;"
            v-permission="'file:create'"
          >
            <el-button type="success">上传文件</el-button>
          </el-upload>
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
        create-permission="file:create"
        update-permission="file:update"
        delete-permission="file:delete"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        @onclick-delete="handleDelete"
        @onclick-update-row="handleUpdate"
        @onclick-add-row="handleAdd"
      >
        <!-- 自定义状态列和操作列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'status'" :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? t('status.enabled') : t('status.disabled') }}
          </el-tag>
        </template>
        <!-- 自定义操作列 -->
        <template #action="{row}">
          <el-button
            type="primary"
            link
            size="small"
            v-permission="'file:list'"
            @click="handleDownload(row)"
          >
            下载
          </el-button>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import {fileList, fileCreate, fileUpdate, fileDelete, fileDownload} from '@/api/generated/admin';
import type {FileItem, FileCreateReq, FileUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';
import {useUserStore} from '@/stores/user';

const {t} = useI18n();

const query = reactive({
  page: 1,
  pageSize: 10,
  name: ''
});
const list = ref<FileItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 文件上传配置
const uploadUrl = computed(() => {
  return `${import.meta.env.VITE_API_BASE_URL || ''}/api/v1/files/upload`;
});

const userStore = useUserStore();
const uploadHeaders = computed(() => {
  return {
    Authorization: `Bearer ${userStore.token}`
  };
});

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'name', label: t('common.name')},
  {prop: 'status', label: t('common.status'), width: 100},
  {prop: 'createdAt', label: t('common.createdAt'), width: 180, type: D2TableElemType.ConvertTime}
]);

// 详情/编辑抽屉列配置
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'name', label: t('common.name'), type: D2TableElemType.EditInput, required: true},
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
    const resp = await fileList({...query});
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

const handleUpdate = async (row: FileItem) => {
  try {
    await fileUpdate(row as FileUpdateReq);
    ElMessage.success('更新成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败');
  }
};

const handleAdd = async (row: any) => {
  try {
    await fileCreate(row as FileCreateReq);
    ElMessage.success('新增成功');
    loadData();
  } catch (err: any) {
    ElMessage.error(err.message || '新增失败');
  }
};

const handleDelete = (index: number, row: FileItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await fileDelete({id: row.id});
      ElMessage.success(t('common.delete'));
      loadData();
    })
    .catch(() => {});
};

// 文件上传前验证
const beforeUpload = (file: File) => {
  const isValidSize = file.size / 1024 / 1024 < 50; // 50MB
  if (!isValidSize) {
    ElMessage.error('文件大小不能超过 50MB');
    return false;
  }
  return true;
};

// 文件上传成功
const handleUploadSuccess = (response: any) => {
  ElMessage.success('文件上传成功');
  loadData();
};

// 文件上传失败
const handleUploadError = (error: any) => {
  ElMessage.error('文件上传失败：' + (error.message || '未知错误'));
};

// 文件下载
const handleDownload = async (row: FileItem) => {
  try {
    // 调用下载接口获取文件URL
    const resp = await fileDownload({id: row.id});
    
    if (resp.url) {
      // 构建完整URL（如果返回的是相对路径，需要拼接baseUrl）
      let downloadUrl = resp.url;
      const baseUrl = import.meta.env.VITE_API_BASE_URL || '';
      if (baseUrl && !resp.url.startsWith('http://') && !resp.url.startsWith('https://')) {
        // 如果是相对路径，拼接baseUrl
        downloadUrl = `${baseUrl}${resp.url.startsWith('/') ? resp.url : `/${resp.url}`}`;
      } else if (!baseUrl && !resp.url.startsWith('http://') && !resp.url.startsWith('https://')) {
        // 如果没有baseUrl，使用相对路径（通过代理）
        downloadUrl = resp.url.startsWith('/') ? resp.url : `/${resp.url}`;
      }
      
      // 创建下载链接
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = row.name;
      link.target = '_blank';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } else {
      ElMessage.error('下载失败：服务器未返回文件URL');
    }
  } catch (err: any) {
    ElMessage.error('下载失败：' + (err.message || '未知错误'));
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

