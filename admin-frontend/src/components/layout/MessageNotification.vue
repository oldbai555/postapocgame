<template>
  <el-popover
    placement="bottom-end"
    :width="350"
    trigger="click"
    popper-class="message-notification-popover"
    @show="handlePopoverShow"
  >
    <template #reference>
      <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
        <el-button text circle class="app-header__action-btn">
          <el-icon :size="18">
            <Bell />
          </el-icon>
        </el-button>
      </el-badge>
    </template>

    <div class="message-notification">
      <div class="message-notification__header">
        <span class="message-notification__title">消息通知</span>
        <div class="message-notification__actions">
          <el-button text size="small" @click="markAllAsRead">全部已读</el-button>
          <el-button text size="small" @click="clearRead">清除已读</el-button>
        </div>
      </div>

      <div class="message-notification__list">
        <div
          v-for="message in displayMessages"
          :key="message.id"
          class="message-notification__item"
          :class="{ unread: !message.read }"
          @click="handleMessageClick(message)"
        >
          <div class="message-notification__item-icon">
            <el-icon v-if="message.type === 'chat'" :size="20">
              <ChatDotRound />
            </el-icon>
            <el-icon v-else-if="message.type === 'task_progress'" :size="20">
              <Loading />
            </el-icon>
            <el-icon v-else :size="20">
              <Bell />
            </el-icon>
          </div>
          <div class="message-notification__item-content">
            <div class="message-notification__item-title">{{ message.title }}</div>
            <div class="message-notification__item-text">{{ message.content }}</div>
            <div class="message-notification__item-time">{{ formatTime(message.timestamp) }}</div>
          </div>
          <div v-if="!message.read" class="message-notification__item-dot"></div>
        </div>

        <el-empty
          v-if="displayMessages.length === 0"
          description="暂无消息"
          :image-size="80"
        />
      </div>

      <div v-if="displayMessages.length > 0" class="message-notification__footer">
        <el-button text size="small" @click="handleViewAll">查看全部</el-button>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import {computed} from 'vue';
import {useRouter} from 'vue-router';
import {Bell, ChatDotRound, Loading} from '@element-plus/icons-vue';
import {useWebSocketStore, type UnreadMessage, MessageType} from '@/stores/websocket';

const router = useRouter();
const wsStore = useWebSocketStore();

const unreadCount = computed(() => wsStore.unreadCount);
const displayMessages = computed(() => wsStore.unreadMessages.slice(0, 10)); // 最多显示 10 条

const formatTime = (timestamp: number) => {
  const now = Date.now();
  const diff = now - timestamp;
  const minutes = Math.floor(diff / 60000);

  if (minutes < 1) {
    return '刚刚';
  } else if (minutes < 60) {
    return `${minutes}分钟前`;
  } else {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
};

const handlePopoverShow = () => {
  // 弹出框显示时，可以做一些处理
};

const handleMessageClick = (message: UnreadMessage) => {
  wsStore.markAsRead(message.id);

  // 根据消息类型跳转
  if (message.type === MessageType.CHAT) {
    router.push('/temp/chat');
  }
};

const markAllAsRead = () => {
  wsStore.markAllAsRead();
};

const clearRead = () => {
  wsStore.clearReadMessages();
};

const handleViewAll = () => {
  router.push('/temp/chat');
};
</script>

<style scoped lang="scss">
.message-notification {
  &__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-light);

    &-title {
      font-weight: 600;
      font-size: 16px;
    }

    &-actions {
      display: flex;
      gap: 8px;
    }
  }

  &__list {
    max-height: 400px;
    overflow-y: auto;
  }

  &__item {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
    cursor: pointer;
    transition: background-color 0.2s;
    position: relative;

    &:hover {
      background-color: var(--el-fill-color-light);
    }

    &.unread {
      background-color: var(--el-color-primary-light-9);
    }

    &-icon {
      flex-shrink: 0;
      color: var(--el-color-primary);
    }

    &-content {
      flex: 1;
      min-width: 0;
    }

    &-title {
      font-weight: 500;
      font-size: 14px;
      margin-bottom: 4px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    &-text {
      font-size: 12px;
      color: var(--el-text-color-secondary);
      margin-bottom: 4px;
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
    }

    &-time {
      font-size: 11px;
      color: var(--el-text-color-placeholder);
    }

    &-dot {
      position: absolute;
      top: 16px;
      right: 16px;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background-color: var(--el-color-primary);
    }
  }

  &__footer {
    padding: 8px 16px;
    border-top: 1px solid var(--el-border-color-light);
    text-align: center;
  }
}
</style>

