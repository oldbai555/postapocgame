<template>
  <div class="page">
    <!-- 系统统计卡片 -->
    <el-row :gutter="20" class="mb-12">
      <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="stat in statsCards" :key="stat.key">
        <el-card class="stat-card">
          <div class="stat-card__content">
            <div class="stat-card__icon" :style="{ backgroundColor: stat.color }">
              <el-icon :size="24">
                <component :is="stat.icon" />
              </el-icon>
            </div>
            <div class="stat-card__info">
              <div class="stat-card__label">{{ stat.label }}</div>
              <div class="stat-card__value">{{ formatNumber(stat.value) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 系统资源监控 -->
    <el-row :gutter="20">
      <!-- CPU 使用率 -->
      <el-col :xs="24" :sm="12" :md="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>CPU 使用率</span>
              <el-button text type="primary" @click="loadMonitorStatus">刷新</el-button>
            </div>
          </template>
          <div class="monitor-item">
            <div class="monitor-item__label">使用率</div>
            <el-progress
              :percentage="Math.round(monitorStatus.cpu.usage)"
              :color="getProgressColor(monitorStatus.cpu.usage)"
              :stroke-width="20"
            />
            <div class="monitor-item__info">
              <span>核心数: {{ monitorStatus.cpu.cores }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 内存使用率 -->
      <el-col :xs="24" :sm="12" :md="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>内存使用率</span>
              <el-button text type="primary" @click="loadMonitorStatus">刷新</el-button>
            </div>
          </template>
          <div class="monitor-item">
            <div class="monitor-item__label">使用率</div>
            <el-progress
              :percentage="Math.round(monitorStatus.memory.usage)"
              :color="getProgressColor(monitorStatus.memory.usage)"
              :stroke-width="20"
            />
            <div class="monitor-item__info">
              <span>已用: {{ formatBytes(monitorStatus.memory.used) }}</span>
              <span>可用: {{ formatBytes(monitorStatus.memory.available) }}</span>
              <span>总计: {{ formatBytes(monitorStatus.memory.total) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 磁盘使用率 -->
      <el-col :xs="24" :sm="12" :md="12" class="mt-20">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>磁盘使用率</span>
              <el-button text type="primary" @click="loadMonitorStatus">刷新</el-button>
            </div>
          </template>
          <div class="monitor-item">
            <div class="monitor-item__label">使用率</div>
            <el-progress
              :percentage="Math.round(monitorStatus.disk.usage)"
              :color="getProgressColor(monitorStatus.disk.usage)"
              :stroke-width="20"
            />
            <div class="monitor-item__info">
              <span>已用: {{ formatBytes(monitorStatus.disk.used) }}</span>
              <span>可用: {{ formatBytes(monitorStatus.disk.available) }}</span>
              <span>总计: {{ formatBytes(monitorStatus.disk.total) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 网络流量 -->
      <el-col :xs="24" :sm="12" :md="12" class="mt-20">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>网络流量</span>
              <el-button text type="primary" @click="loadMonitorStatus">刷新</el-button>
            </div>
          </template>
          <div class="monitor-item">
            <div class="monitor-item__network">
              <div class="network-item">
                <div class="network-item__label">发送</div>
                <div class="network-item__value">{{ formatBytes(monitorStatus.network.bytesSent) }}</div>
                <div class="network-item__label">包数: {{ formatNumber(monitorStatus.network.packetsSent) }}</div>
              </div>
              <div class="network-item">
                <div class="network-item__label">接收</div>
                <div class="network-item__value">{{ formatBytes(monitorStatus.network.bytesRecv) }}</div>
                <div class="network-item__label">包数: {{ formatNumber(monitorStatus.network.packetsRecv) }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, onUnmounted, computed} from 'vue';
import {ElMessage} from 'element-plus';
import {User, Key, Menu, Document, Connection} from '@element-plus/icons-vue';
import {monitorStatus as monitorStatusApi, monitorStats as monitorStatsApi} from '@/api/generated/admin';
import type {MonitorStatusResp, MonitorStatsResp} from '@/api/generated/admin';

const loading = ref(false);
const monitorStatusData = ref<MonitorStatusResp>({
  cpu: {usage: 0, cores: 0},
  memory: {total: 0, used: 0, available: 0, usage: 0},
  disk: {total: 0, used: 0, available: 0, usage: 0},
  network: {bytesSent: 0, bytesRecv: 0, packetsSent: 0, packetsRecv: 0}
});
const monitorStatsData = ref<MonitorStatsResp>({
  userCount: 0,
  roleCount: 0,
  permissionCount: 0,
  menuCount: 0,
  onlineUserCount: 0,
  operationLogCount: 0,
  loginLogCount: 0
});

// 统计卡片数据
const statsCards = computed(() => [
  {key: 'user', label: '用户总数', value: monitorStatsData.value.userCount, icon: User, color: '#409EFF'},
  {key: 'role', label: '角色总数', value: monitorStatsData.value.roleCount, icon: Key, color: '#67C23A'},
  {key: 'permission', label: '权限总数', value: monitorStatsData.value.permissionCount, icon: Key, color: '#E6A23C'},
  {key: 'menu', label: '菜单总数', value: monitorStatsData.value.menuCount, icon: Menu, color: '#F56C6C'},
  {key: 'online', label: '在线用户', value: monitorStatsData.value.onlineUserCount, icon: Connection, color: '#909399'},
  {key: 'operationLog', label: '操作日志', value: monitorStatsData.value.operationLogCount, icon: Document, color: '#606266'},
  {key: 'loginLog', label: '登录日志', value: monitorStatsData.value.loginLogCount, icon: Document, color: '#909399'}
]);

// 监控状态数据（用于显示）
const monitorStatus = computed(() => monitorStatusData.value);

// 加载监控状态
const loadMonitorStatus = async () => {
  loading.value = true;
  try {
    const resp = await monitorStatusApi();
    monitorStatusData.value = resp;
  } catch (err: any) {
    ElMessage.error(err.message || '获取监控状态失败');
  } finally {
    loading.value = false;
  }
};

// 加载统计数据
const loadMonitorStats = async () => {
  try {
    const resp = await monitorStatsApi();
    monitorStatsData.value = resp;
  } catch (err: any) {
    ElMessage.error(err.message || '获取统计数据失败');
  }
};

// 格式化数字
const formatNumber = (num: number) => {
  return num.toLocaleString();
};

// 格式化字节数
const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
};

// 获取进度条颜色
const getProgressColor = (percentage: number) => {
  if (percentage < 50) return '#67c23a';
  if (percentage < 80) return '#e6a23c';
  return '#f56c6c';
};

// 定时刷新
let refreshTimer: number | null = null;

// 初始化加载
onMounted(() => {
  loadMonitorStatus();
  loadMonitorStats();
  // 每30秒自动刷新
  refreshTimer = window.setInterval(() => {
    loadMonitorStatus();
    loadMonitorStats();
  }, 30000);
});

// 清理定时器
onUnmounted(() => {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer);
  }
});
</script>

<style scoped lang="scss">
.page {
  padding: 20px;
}

.mb-12 {
  margin-bottom: 12px;
}

.mt-20 {
  margin-top: 20px;
}

.stat-card {
  margin-bottom: 20px;
  
  &__content {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  
  &__icon {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
  }
  
  &__info {
    flex: 1;
  }
  
  &__label {
    font-size: 14px;
    color: #909399;
    margin-bottom: 4px;
  }
  
  &__value {
    font-size: 24px;
    font-weight: bold;
    color: #303133;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.monitor-item {
  &__label {
    font-size: 14px;
    color: #606266;
    margin-bottom: 8px;
  }
  
  &__info {
    display: flex;
    gap: 16px;
    margin-top: 8px;
    font-size: 12px;
    color: #909399;
  }
  
  &__network {
    display: flex;
    gap: 24px;
  }
}

.network-item {
  flex: 1;
  
  &__label {
    font-size: 12px;
    color: #909399;
    margin-bottom: 4px;
  }
  
  &__value {
    font-size: 18px;
    font-weight: bold;
    color: #303133;
    margin-bottom: 4px;
  }
}
</style>

