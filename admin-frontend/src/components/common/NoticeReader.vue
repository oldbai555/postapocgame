<template>
  <el-dialog
    :model-value="visible"
    title="公告阅读"
    width="600px"
    :close-on-click-modal="false"
    @update:model-value="handleVisibleChange"
    @close="handleClose"
  >
    <div v-if="currentNotice" class="notice-reader">
      <div class="notice-reader__header">
        <h3 class="notice-reader__title">{{ currentNotice.title }}</h3>
        <div class="notice-reader__meta">
          <el-tag :type="getNoticeTypeTag(currentNotice.type)" size="small">
            {{ getNoticeTypeLabel(currentNotice.type) }}
          </el-tag>
          <span class="notice-reader__time">{{ formatTime(currentNotice.publishTime) }}</span>
        </div>
      </div>
      <div class="notice-reader__content">
        <div class="notice-reader__text" v-html="formatContent(currentNotice.content)"></div>
      </div>
    </div>

    <template #footer>
      <div class="notice-reader__footer">
        <div class="notice-reader__nav">
          <el-button
            :disabled="currentIndex === 0"
            @click="handlePrev"
          >
            上一条
          </el-button>
          <span class="notice-reader__counter">
            {{ currentIndex + 1 }} / {{ notices.length }}
          </span>
          <el-button
            :disabled="currentIndex >= notices.length - 1"
            @click="handleNext"
          >
            下一条
          </el-button>
        </div>
        <el-button type="primary" @click="handleMarkAsRead">
          已读
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue';
import {ElMessage} from 'element-plus';
import type {NotificationItem} from '@/api/generated/admin';
import {notificationRead} from '@/api/generated/admin';

interface Props {
  notices: NotificationItem[];
  visible: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  'update:visible': [value: boolean];
  'read': [id: number];
}>();

const currentIndex = ref(0);

const currentNotice = computed(() => {
  if (props.notices.length === 0) return null;
  const notification = props.notices[currentIndex.value];
  if (!notification) return null;
  
  // 从通知中获取公告信息（需要根据 sourceId 查询公告详情）
  // 这里暂时使用通知的 title 和 content
  return {
    id: notification.sourceId,
    title: notification.title,
    content: notification.content,
    type: 1, // 默认普通公告，实际应该从公告详情获取
    publishTime: notification.createdAt || 0 // createdAt 已经是秒级时间戳
  };
});

const getNoticeTypeLabel = (type: number): string => {
  const map: Record<number, string> = {
    1: '普通公告',
    2: '重要公告',
    3: '紧急公告'
  };
  return map[type] || '未知';
};

const getNoticeTypeTag = (type: number): string => {
  const map: Record<number, string> = {
    1: 'info',
    2: 'warning',
    3: 'danger'
  };
  return map[type] || '';
};

const formatTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  const date = new Date(timestamp * 1000);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
};

const formatContent = (content: string): string => {
  if (!content) return '';
  // 简单的换行处理
  return content.replace(/\n/g, '<br>');
};

const handlePrev = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--;
  }
};

const handleNext = () => {
  if (currentIndex.value < props.notices.length - 1) {
    currentIndex.value++;
  }
};

const handleMarkAsRead = async () => {
  if (!currentNotice.value) return;
  
  const notification = props.notices[currentIndex.value];
  if (!notification) return;

  try {
    // 标记当前通知为已读
    await notificationRead({id: notification.id});
    emit('read', notification.id);
    
    // 如果还有下一条，自动切换到下一条
    if (currentIndex.value < props.notices.length - 1) {
      currentIndex.value++;
    } else {
      // 如果没有下一条了，关闭对话框
      emit('update:visible', false);
    }
  } catch (err: any) {
    ElMessage.error(err.message || '标记已读失败');
  }
};

const handleVisibleChange = (value: boolean) => {
  emit('update:visible', value);
};

const handleClose = () => {
  emit('update:visible', false);
};

// 当 visible 变为 true 时，重置索引
watch(() => props.visible, (newVal) => {
  if (newVal) {
    currentIndex.value = 0;
  }
});
</script>

<style scoped lang="scss">
.notice-reader {
  &__header {
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--el-border-color-light);
  }

  &__title {
    margin: 0 0 12px 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  &__meta {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__time {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  &__content {
    min-height: 200px;
    max-height: 400px;
    overflow-y: auto;
    padding: 16px;
    background-color: var(--el-fill-color-lighter);
    border-radius: 4px;
  }

  &__text {
    line-height: 1.8;
    color: var(--el-text-color-primary);
    white-space: pre-wrap;
    word-break: break-word;
  }

  &__footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  &__nav {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__counter {
    font-size: 14px;
    color: var(--el-text-color-secondary);
    min-width: 60px;
    text-align: center;
  }
}
</style>

