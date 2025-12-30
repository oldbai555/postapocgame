<template>
  <div class="chat-container">
    <el-card class="chat-card">
      <template #header>
        <div class="chat-header">
          <span class="chat-title">在线聊天</span>
          <div class="chat-status">
            <el-tag :type="wsConnected ? 'success' : 'danger'" size="small">
              {{ wsConnected ? '已连接' : '未连接' }}
            </el-tag>
            <el-button
              v-if="!wsConnected"
              type="primary"
              size="small"
              @click="wsStore.connect()"
            >
              重新连接
            </el-button>
          </div>
        </div>
      </template>

      <div class="chat-content">
        <!-- 左侧：在线用户列表 -->
        <div class="chat-sidebar">
          <div class="sidebar-header">
            <h3>在线用户 ({{ onlineUsers.length }})</h3>
          </div>
          <div class="user-list">
            <div
              v-for="user in sortedOnlineUsers"
              :key="user.userId"
              class="user-item"
              :class="{ active: selectedUserId === user.userId, 'is-me': Number(user.userId) === Number(currentUserId) }"
              @click="selectUser(user.userId)"
            >
              <el-avatar :size="32" :src="user.avatar || ''">
                {{ user.userName?.charAt(0).toUpperCase() || 'U' }}
              </el-avatar>
              <span class="username">{{ user.userName }}</span>
              <el-tag v-if="Number(user.userId) === Number(currentUserId)" size="small" type="info">我</el-tag>
            </div>
            <div
              v-if="onlineUsers.length === 0"
              class="empty-users"
            >
              <el-empty description="暂无在线用户" :image-size="80" />
            </div>
          </div>
        </div>

        <!-- 右侧：聊天区域 -->
        <div class="chat-main">
          <!-- 消息列表 -->
          <div class="message-list" ref="messageListRef">
            <div
              v-for="message in messages"
              :key="message.id"
              class="message-item"
              :class="{ 'message-self': Number(message.fromUserId) === Number(currentUserId) }"
            >
              <div class="message-avatar">
                <el-avatar :size="36">
                  {{ message.fromUserName?.charAt(0).toUpperCase() || 'U' }}
                </el-avatar>
              </div>
              <div class="message-content">
                <div class="message-header">
                  <span class="message-username">{{ message.fromUserName }}</span>
                  <span class="message-time">{{ formatTime(message.createdAt) }}</span>
                </div>
                <div class="message-text">{{ message.content }}</div>
              </div>
            </div>
            <div v-if="messages.length === 0" class="empty-message">
              <el-empty description="暂无消息，开始聊天吧~" />
            </div>
          </div>

          <!-- 输入区域 -->
          <div class="message-input">
            <el-input
              v-model="inputMessage"
              type="textarea"
              :rows="3"
              placeholder="输入消息..."
              @keydown.enter.exact.prevent="handleSendMessage"
              @keydown.enter.shift.exact="inputMessage += '\n'"
            />
            <div class="input-actions">
              <div class="input-info">
                <span v-if="!selectedUserId">房间: {{ currentRoomId }}</span>
                <span v-else>私聊: {{ selectedUsername }}</span>
              </div>
              <el-button
                type="primary"
                :disabled="!inputMessage.trim() || !wsConnected"
                @click="handleSendMessage"
              >
                发送
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { useRoute } from 'vue-router';
import { useUserStore } from '@/stores/user';
import { useWebSocketStore, MessageType } from '@/stores/websocket';
import {
  chatMessageList,
  chatMessageSend,
  chatOnlineUsers,
  dictGet
} from '@/api/generated/admin';
import type {
  ChatMessageItem,
  ChatMessageListReq,
  ChatMessageSendReq,
  ChatOnlineUserItem
} from '@/api/generated/admin';

const route = useRoute();
const userStore = useUserStore();
const wsStore = useWebSocketStore();
const currentUserId = computed(() => userStore.profile?.id || 0);
const currentUsername = computed(() => userStore.profile?.username || '');

// WebSocket 连接状态（使用全局 store）
const wsConnected = computed(() => wsStore.connected);

// 聊天相关
const currentRoomId = ref('default');
const selectedUserId = ref<number | null>(null);
const selectedUsername = ref('');
const inputMessage = ref('');
const messages = ref<ChatMessageItem[]>([]);
const onlineUsers = ref<ChatOnlineUserItem[]>([]);
const messageListRef = ref<HTMLElement>();

// 排序后的在线用户列表（"我"固定在第一位）
const sortedOnlineUsers = computed(() => {
  const users = [...onlineUsers.value];
  const myIndex = users.findIndex((u) => Number(u.userId) === Number(currentUserId.value));
  
  if (myIndex > 0) {
    // 如果"我"不在第一位，将其移到第一位
    const me = users.splice(myIndex, 1)[0];
    users.unshift(me);
  }
  
  return users;
});

// 聊天配置：消息数量限制（从字典获取，默认30）
const chatMessageLimit = ref(30);

// 查询参数
const query = reactive<ChatMessageListReq>({
  page: 1,
  pageSize: 30, // 默认30，将从字典加载后更新
  roomId: currentRoomId.value,
  userId: 0
});

// 监听 WebSocket 消息（使用全局 store）
watch(
  () => wsStore.lastMessage,
  (newMessage) => {
    if (!newMessage) return;

    // 只处理聊天相关的消息
    if (newMessage.type === MessageType.CHAT || newMessage.type === 'chat') {
      handleChatMessage(newMessage);
    } else if (newMessage.type === 'join') {
      ElMessage.info(`${newMessage.fromName} 加入了聊天室`);
      loadOnlineUsers();
    } else if (newMessage.type === 'leave') {
      ElMessage.info(`${newMessage.fromName} 离开了聊天室`);
      loadOnlineUsers();
    }
  }
);

// 处理聊天消息
const handleChatMessage = (data: any) => {
  // 优先判断是否是私聊消息（toId > 0 表示私聊）
  const isPrivateMessage = data.toId > 0;
  
  if (isPrivateMessage) {
    // 私聊消息：只显示在对应的私聊窗口中
    // 检查是否是当前选中的私聊对象
    if (selectedUserId.value && 
        ((Number(data.fromId) === Number(selectedUserId.value) && Number(data.toId) === Number(currentUserId.value)) ||
         (Number(data.toId) === Number(selectedUserId.value) && Number(data.fromId) === Number(currentUserId.value)))) {
      // 收到新消息，添加到消息列表
      const newMessage: ChatMessageItem = {
        id: data.messageId || Date.now(),
        fromUserId: data.fromId,
        fromUserName: data.fromName,
        toUserId: data.toId || 0,
        toUserName: '',
        roomId: data.roomId || '',
        content: data.content,
        messageType: 1,
        status: 1,
        createdAt: data.createdAt || new Date().toISOString()
      };
      messages.value.push(newMessage);
      // 如果消息数量超过限制，只保留最新的N条
      if (messages.value.length > chatMessageLimit.value) {
        messages.value = messages.value.slice(-chatMessageLimit.value);
      }
      scrollToBottom();
    }
    // 如果不在私聊窗口，不添加到消息列表（避免私聊消息出现在群聊中）
  } else {
    // 群聊消息：只显示在当前房间
    const isCurrentRoom = !selectedUserId.value && data.roomId === currentRoomId.value;
    if (isCurrentRoom) {
      // 收到新消息，添加到消息列表
      const newMessage: ChatMessageItem = {
        id: data.messageId || Date.now(),
        fromUserId: data.fromId,
        fromUserName: data.fromName,
        toUserId: data.toId || 0,
        toUserName: '',
        roomId: data.roomId || currentRoomId.value,
        content: data.content,
        messageType: 1,
        status: 1,
        createdAt: data.createdAt || new Date().toISOString()
      };
      messages.value.push(newMessage);
      // 如果消息数量超过限制，只保留最新的N条
      if (messages.value.length > chatMessageLimit.value) {
        messages.value = messages.value.slice(-chatMessageLimit.value);
      }
      scrollToBottom();
    }
  }
};

// 发送消息
const handleSendMessage = async () => {
  if (!inputMessage.value.trim()) {
    return;
  }

  if (!wsConnected.value) {
    ElMessage.warning('WebSocket 未连接，请先连接');
    return;
  }

  const messageContent = inputMessage.value.trim();
  inputMessage.value = '';

  try {
    const req: ChatMessageSendReq = {
      toUserId: selectedUserId.value || 0,
      roomId: currentRoomId.value,
      content: messageContent,
      messageType: 1
    };

    await chatMessageSend(req);
    // 消息会通过 WebSocket 推送回来，不需要手动添加到列表
  } catch (err: any) {
    ElMessage.error(err.message || '发送消息失败');
    // 恢复输入内容
    inputMessage.value = messageContent;
  }
};

// 加载聊天配置（从字典获取）
const loadChatConfig = async () => {
  try {
    const resp = await dictGet({code: 'chat_config'});
    if (resp && resp.items && resp.items.length > 0) {
      // 查找"聊天窗口消息数量"配置项
      const limitItem = resp.items.find(item => item.label === '聊天窗口消息数量');
      if (limitItem && limitItem.value) {
        const limit = parseInt(limitItem.value, 10);
        if (!isNaN(limit) && limit > 0) {
          chatMessageLimit.value = limit;
          query.pageSize = limit;
        }
      }
    }
  } catch (err: any) {
    console.warn('加载聊天配置失败，使用默认值:', err);
    // 使用默认值30
    chatMessageLimit.value = 30;
    query.pageSize = 30;
  }
};

// 加载消息列表
const loadMessages = async () => {
  try {
    // 重置查询参数
    query.page = 1;
    query.pageSize = chatMessageLimit.value; // 使用从字典获取的限制值
    
    if (selectedUserId.value) {
      // 私聊：查询与当前用户和选中用户之间的消息
      query.roomId = ''; // 私聊时不使用 roomId
      query.userId = selectedUserId.value;
    } else {
      // 群聊：查询房间内的消息
      query.roomId = currentRoomId.value;
      query.userId = 0; // 群聊时 userId 为 0
    }
    
    const resp = await chatMessageList(query);
    const allMessages = (resp.list || []).reverse(); // 反转列表，最新的在底部
    // 只保留最新的N条消息（N为字典配置的值）
    messages.value = allMessages.slice(-chatMessageLimit.value);
    nextTick(() => {
      scrollToBottom();
    });
  } catch (err: any) {
    ElMessage.error(err.message || '加载消息失败');
  }
};

// 加载在线用户列表
const loadOnlineUsers = async () => {
  try {
    const resp = await chatOnlineUsers();
    onlineUsers.value = resp.list || [];
  } catch (err: any) {
    console.error('加载在线用户失败:', err);
  }
};

// 选择用户（私聊）
const selectUser = (userId: number) => {
  if (userId === currentUserId.value) {
    selectedUserId.value = null;
    selectedUsername.value = '';
    loadMessages();
    return;
  }
  selectedUserId.value = userId;
  const user = onlineUsers.value.find((u) => u.userId === userId);
  selectedUsername.value = user?.userName || '';
  loadMessages();
};

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight;
    }
  });
};

// 格式化时间（接受秒级时间戳）
const formatTime = (timestamp: number) => {
  if (!timestamp) return '';
  const date = new Date(timestamp * 1000); // 秒级时间戳转换为毫秒
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);

  if (minutes < 1) {
    return '刚刚';
  } else if (minutes < 60) {
    return `${minutes}分钟前`;
  } else if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  } else {
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
};

onMounted(async () => {
  // 确保用户信息已加载（刷新页面后可能需要重新获取）
  if (userStore.token && !userStore.profile) {
    await userStore.fetchProfile(true);
  }
  
  // 先加载聊天配置，再加载消息
  await loadChatConfig();
  loadMessages();
  loadOnlineUsers();
  // WebSocket 连接由全局 store 管理，这里不需要手动连接
  // 但确保连接已建立
  if (!wsConnected.value && userStore.token) {
    wsStore.connect();
  }
});

onUnmounted(() => {
  // 不断开 WebSocket，因为可能在其他页面还需要使用
});
</script>

<style scoped lang="scss">
.chat-container {
  height: calc(100vh - 120px);
  padding: 20px;

  .chat-card {
    height: 100%;

    :deep(.el-card__body) {
      height: calc(100% - 60px);
      padding: 0;
    }
  }
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .chat-title {
    font-size: 18px;
    font-weight: 600;
  }

  .chat-status {
    display: flex;
    align-items: center;
    gap: 10px;
  }
}

.chat-content {
  display: flex;
  height: 100%;
  border-top: 1px solid var(--el-border-color-light);
}

.chat-sidebar {
  width: 250px;
  border-right: 1px solid var(--el-border-color-light);
  display: flex;
  flex-direction: column;

  .sidebar-header {
    padding: 15px;
    border-bottom: 1px solid var(--el-border-color-light);

    h3 {
      margin: 0;
      font-size: 14px;
      font-weight: 600;
    }
  }

  .user-list {
    flex: 1;
    overflow-y: auto;
    padding: 10px;

    .user-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px;
      border-radius: 6px;
      cursor: pointer;
      transition: background-color 0.2s;

      &:hover {
        background-color: var(--el-fill-color-light);
      }

      &.active {
        background-color: var(--el-color-primary-light-9);
      }

      &.is-me {
        opacity: 0.7;
      }

      .username {
        flex: 1;
        font-size: 14px;
      }
    }

    .empty-users {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 200px;
    }
  }
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: var(--el-bg-color-page);

  .message-item {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 20px;

    &.message-self {
      flex-direction: row-reverse;

      .message-content {
        align-items: flex-end;

        .message-header {
          flex-direction: row-reverse;
        }

        .message-text {
          background-color: var(--el-color-primary);
          color: white;
        }
      }
    }

    .message-avatar {
      flex-shrink: 0;
    }

    .message-content {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 6px;
      max-width: 60%;

      .message-header {
        display: flex;
        align-items: center;
        gap: 8px;

        .message-username {
          font-size: 13px;
          font-weight: 600;
          color: var(--el-text-color-primary);
        }

        .message-time {
          font-size: 12px;
          color: var(--el-text-color-secondary);
        }
      }

      .message-text {
        padding: 10px 14px;
        border-radius: 8px;
        background-color: var(--el-bg-color);
        font-size: 14px;
        line-height: 1.5;
        word-wrap: break-word;
        white-space: pre-wrap;
      }
    }
  }

  .empty-message {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
  }
}

.message-input {
  padding: 15px;
  border-top: 1px solid var(--el-border-color-light);
  background-color: var(--el-bg-color);

  .input-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 10px;

    .input-info {
      display: flex;
      gap: 15px;
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }
}
</style>
