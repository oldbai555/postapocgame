<template>
  <div class="page">
    <!-- 搜索表单 -->
    <el-card class="mb-12">
      <el-form :inline="true" :model="query">
        <el-form-item label="Method">
          <el-input v-model="query.method" placeholder="HTTP Method" clearable />
        </el-form-item>
        <el-form-item label="Path">
          <el-input v-model="query.path" placeholder="API Path" clearable />
        </el-form-item>
        <el-form-item label="Status">
          <el-input v-model.number="query.statusCode" placeholder="HTTP Status Code" clearable />
        </el-form-item>
        <el-form-item label="Slow Flag">
          <el-select v-model="query.isSlow" placeholder="Slow or not" clearable>
            <el-option label="Slow" :value="1" />
            <el-option label="Normal" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="Start Time">
          <el-date-picker
            v-model="query.startTime"
            type="datetime"
            placeholder="Start Time"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            clearable
          />
        </el-form-item>
        <el-form-item label="End Time">
          <el-date-picker
            v-model="query.endTime"
            type="datetime"
            placeholder="End Time"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">Search</el-button>
          <el-button @click="handleReset">Reset</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- D2Table 组件（只读列表） -->
    <el-card>
      <D2Table
        :columns="columns"
        :data="list"
        :total="total"
        :page-size="query.pageSize"
        :current-page="query.page"
        :drawer-columns="drawerColumns"
        :drawer-add-columns="drawerAddColumns"
        :have-edit="false"
        :have-detail="false"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      >
        <!-- 自定义慢接口标记列 -->
        <template #cell="{row, column}">
          <el-tag v-if="column.prop === 'isSlow'" :type="row.isSlow === 1 ? 'danger' : 'info'">
            {{ row.isSlow === 1 ? 'Slow' : 'Normal' }}
          </el-tag>
        </template>
      </D2Table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref, onMounted, computed} from 'vue';
import {ElMessage} from 'element-plus';
import { performanceLogList } from '@/api/generated/admin';
import type { PerformanceLogItem, PerformanceLogListReq } from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';
import D2Table from '@/components/common/D2Table.vue';
import {D2TableElemType, type TableColumn, type DrawerColumn} from '@/types/table';

const {t} = useI18n();

const query = reactive<PerformanceLogListReq & { page: number; pageSize: number }>({
  page: 1,
  pageSize: 10,
  method: '',
  path: '',
  isSlow: undefined,
  statusCode: undefined,
  startTime: '',
  endTime: ''
});
const list = ref<PerformanceLogItem[]>([]);
const total = ref(0);
const loading = ref(false);

// 表格列配置（只读性能日志字段）
const columns = computed<TableColumn[]>(() => [
  {prop: 'id', label: 'ID', width: 80},
  {prop: 'method', label: 'Method', width: 90},
  {prop: 'path', label: 'Path', minWidth: 220, showOverflowTooltip: true},
  {prop: 'statusCode', label: 'Status', width: 90},
  {prop: 'duration', label: 'Duration (ms)', width: 120},
  {prop: 'isSlow', label: 'Slow Flag', width: 100},
  {prop: 'username', label: t('common.username'), width: 140},
  {prop: 'ipAddress', label: 'IP', width: 140},
  {prop: 'createdAt', label: t('common.createdAt'), width: 180}
]);

// 占位：只读模式但 D2Table 要求必传
const drawerColumns: DrawerColumn[] = [];
const drawerAddColumns: DrawerColumn[] = [];

const loadData = async () => {
  loading.value = true;
  try {
    const resp = await performanceLogList({...query});
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
  query.method = '';
  query.path = '';
  query.isSlow = undefined;
  query.statusCode = undefined;
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


