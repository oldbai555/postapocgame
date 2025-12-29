<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="用户ID">
          <el-input v-model.number="query.userId" placeholder="用户ID" clearable />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="query.username" placeholder="用户名" clearable />
        </el-form-item>
        <el-form-item label="审计类型">
          <el-select v-model="query.auditType" placeholder="审计类型" clearable>
            <el-option label="权限分配" value="permission_assign" />
            <el-option label="角色变更" value="role_change" />
            <el-option label="配置修改" value="config_modify" />
            <el-option label="数据删除" value="data_delete" />
          </el-select>
        </el-form-item>
        <el-form-item label="审计对象">
          <el-input v-model="query.auditObject" placeholder="审计对象" clearable />
        </el-form-item>
        <el-form-item label="开始时间">
          <el-date-picker
            v-model="query.startTime"
            type="datetime"
            placeholder="开始时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            clearable
          />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-date-picker
            v-model="query.endTime"
            type="datetime"
            placeholder="结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleExport" v-permission="'audit_log:export'">导出</el-button>
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
        :have-edit="false"
        :have-detail="true"
        detail-permission="audit_log:detail"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      >
        <!-- 自定义审计类型列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'auditType'" :type="getAuditTypeTagType(row.auditType)">
            {{ getAuditTypeLabel(row.auditType) }}
          </el-tag>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage} from 'element-plus';
import { auditLogList } from '@/api/generated/admin';
import type { AuditLogItem, AuditLogListReq, AuditLogExportReq } from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';
import request from '@/utils/request';

const {t} = useI18n();

const query = reactive<AuditLogListReq>({
  page: 1,
  pageSize: 20,
  userId: undefined,
  username: '',
  auditType: '',
  auditObject: '',
  startTime: '',
  endTime: ''
});
const list = ref<AuditLogItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 审计类型标签映射
const getAuditTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    permission_assign: '权限分配',
    role_change: '角色变更',
    config_modify: '配置修改',
    data_delete: '数据删除'
  };
  return map[type] || type;
};

const getAuditTypeTagType = (type: string) => {
  const map: Record<string, string> = {
    permission_assign: 'success',
    role_change: 'warning',
    config_modify: 'info',
    data_delete: 'danger'
  };
  // 默认统一用 info，避免传入非法的空字符串导致 Element Plus 报错
  return map[type] || 'info';
};

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'userId', label: '用户ID', width: 100},
  {prop: 'username', label: '用户名', width: 120},
  {prop: 'auditType', label: '审计类型', width: 120},
  {prop: 'auditObject', label: '审计对象', width: 150},
  {prop: 'ipAddress', label: 'IP地址', width: 140},
  {prop: 'createdAt', label: '创建时间', width: 180, type: D2TableElemType.ConvertTime}
]);

// 详情抽屉列配置（只读）
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'userId', label: '用户ID', type: D2TableElemType.Tag},
  {prop: 'username', label: '用户名', type: D2TableElemType.Tag},
  {prop: 'auditType', label: '审计类型', type: D2TableElemType.Tag},
  {prop: 'auditObject', label: '审计对象', type: D2TableElemType.Tag},
  {prop: 'auditDetail', label: '审计详情', type: D2TableElemType.Textarea},
  {prop: 'ipAddress', label: 'IP地址', type: D2TableElemType.Tag},
  {prop: 'userAgent', label: '用户代理', type: D2TableElemType.Textarea},
  {prop: 'createdAt', label: '创建时间', type: D2TableElemType.ConvertTime}
]);

const loadData = async () => {
  loading.value = true;
  try {
    const req: AuditLogListReq = {
      page: query.page,
      pageSize: query.pageSize,
      userId: query.userId,
      username: query.username || undefined,
      auditType: query.auditType || undefined,
      auditObject: query.auditObject || undefined,
      startTime: query.startTime || undefined,
      endTime: query.endTime || undefined
    };
    const resp = await auditLogList(req);
    list.value = resp.list;
    total.value = resp.total;
  } catch (err: any) {
    ElMessage.error(err.message || '查询失败');
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  query.page = 1;
  query.pageSize = 20;
  query.userId = undefined;
  query.username = '';
  query.auditType = '';
  query.auditObject = '';
  query.startTime = '';
  query.endTime = '';
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

const handleExport = async () => {
  try {
    const params: any = {};
    if (query.userId) params.userId = query.userId;
    if (query.username) params.username = query.username;
    if (query.auditType) params.auditType = query.auditType;
    if (query.auditObject) params.auditObject = query.auditObject;
    if (query.startTime) params.startTime = query.startTime;
    if (query.endTime) params.endTime = query.endTime;
    
    // 使用 request 下载文件，设置 responseType 为 blob
    const resp: any = await request.get('/v1/audit-logs/export', {
      params,
      responseType: 'blob'
    });
    
    // 创建 Blob URL（resp 已经是 Blob 类型）
    const blob = resp instanceof Blob ? resp : new Blob([resp], {type: 'text/csv;charset=utf-8'});
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `审计日志_${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    // 释放 Blob URL
    window.URL.revokeObjectURL(url);
    ElMessage.success('导出成功');
  } catch (err: any) {
    ElMessage.error(err.message || '导出失败');
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


