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
        <el-form-item label="操作类型">
          <el-select v-model="query.operationType" placeholder="操作类型" clearable>
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
            <el-option label="查询" value="query" />
            <el-option label="导出" value="export" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作对象">
          <el-input v-model="query.operationObject" placeholder="操作对象" clearable />
        </el-form-item>
        <el-form-item label="请求方法">
          <el-select v-model="query.method" placeholder="请求方法" clearable>
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
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
          <el-button type="success" @click="handleExport">导出</el-button>
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
        detail-permission="operation_log:detail"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      >
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage} from 'element-plus';
import { operationLogList } from '@/api/generated/admin';
import type { OperationLogItem, OperationLogListReq, OperationLogExportReq } from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';
import request from '@/utils/request';

const {t} = useI18n();

const query = reactive<OperationLogListReq>({
  page: 1,
  pageSize: 20,
  userId: undefined,
  username: '',
  operationType: '',
  operationObject: '',
  method: '',
  startTime: '',
  endTime: ''
});
const list = ref<OperationLogItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 表格列配置
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'userId', label: '用户ID', width: 100},
  {prop: 'username', label: '用户名', width: 120},
  {prop: 'operationType', label: '操作类型', width: 100},
  {prop: 'operationObject', label: '操作对象', width: 120},
  {prop: 'method', label: '请求方法', width: 100},
  {prop: 'path', label: '请求路径', minWidth: 200},
  {prop: 'ipAddress', label: 'IP地址', width: 140},
  {prop: 'duration', label: '耗时(ms)', width: 100},
  {prop: 'createdAt', label: '创建时间', width: 180}
]);

// 详情抽屉列配置（只读）
const drawerColumns = computed<DrawerColumn[]>(() => [
  {prop: 'id', label: 'ID', type: D2TableElemType.Tag},
  {prop: 'userId', label: '用户ID', type: D2TableElemType.Tag},
  {prop: 'username', label: '用户名', type: D2TableElemType.Tag},
  {prop: 'operationType', label: '操作类型', type: D2TableElemType.Tag},
  {prop: 'operationObject', label: '操作对象', type: D2TableElemType.Tag},
  {prop: 'method', label: '请求方法', type: D2TableElemType.Tag},
  {prop: 'path', label: '请求路径', type: D2TableElemType.Tag},
  {prop: 'requestParams', label: '请求参数', type: D2TableElemType.Textarea},
  {prop: 'responseCode', label: '响应状态码', type: D2TableElemType.Tag},
  {prop: 'responseMsg', label: '响应消息', type: D2TableElemType.Tag},
  {prop: 'ipAddress', label: 'IP地址', type: D2TableElemType.Tag},
  {prop: 'userAgent', label: '用户代理', type: D2TableElemType.Textarea},
  {prop: 'duration', label: '耗时(ms)', type: D2TableElemType.Tag},
  {prop: 'createdAt', label: '创建时间', type: D2TableElemType.Tag}
]);

const loadData = async () => {
  loading.value = true;
  try {
    const req: OperationLogListReq = {
      page: query.page,
      pageSize: query.pageSize,
      userId: query.userId,
      username: query.username || undefined,
      operationType: query.operationType || undefined,
      operationObject: query.operationObject || undefined,
      method: query.method || undefined,
      startTime: query.startTime || undefined,
      endTime: query.endTime || undefined
    };
    const resp = await operationLogList(req);
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
  query.operationType = '';
  query.operationObject = '';
  query.method = '';
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
    if (query.operationType) params.operationType = query.operationType;
    if (query.operationObject) params.operationObject = query.operationObject;
    if (query.method) params.method = query.method;
    if (query.startTime) params.startTime = query.startTime;
    if (query.endTime) params.endTime = query.endTime;
    
    // 使用 request 下载文件，设置 responseType 为 blob
    const resp: any = await request.get('/v1/operation-logs/export', {
      params,
      responseType: 'blob'
    });
    
    // 创建 Blob URL（resp 已经是 Blob 类型）
    const blob = resp instanceof Blob ? resp : new Blob([resp], {type: 'text/csv;charset=utf-8'});
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `操作日志_${new Date().toISOString().slice(0, 10)}.csv`;
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


