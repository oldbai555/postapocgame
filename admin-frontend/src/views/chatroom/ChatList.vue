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
        <!-- 左侧：聊天列表 -->
        <div class="chat-sidebar">
          <div class="sidebar-header">
            <h3>聊天列表 ({{ chats.length }})</h3>
          </div>
          <div class="user-list">
            <div
                v-for="chat in chats"
                :key="chat.chatId"
                class="user-item"
                :class="{ active: selectedChatId === chat.chatId }"
                @click="selectChat(chat)"
            >
              <el-avatar :size="32" :src="chat.avatar || ''">
                {{ chat.name?.charAt(0).toUpperCase() || 'C' }}
              </el-avatar>
              <div class="user-info">
                <div class="user-name">
                  {{ chat.name }}
                  <el-tag v-if="chat.chatType === 2" size="small" type="info" style="margin-left: 4px">群组</el-tag>
                </div>
                <div class="user-desc" v-if="chat.chatType === 1">
                  {{ formatChatDesc(chat) }}
                </div>
                <div class="user-desc" v-else-if="chat.description">
                  {{ chat.description }}
                </div>
              </div>
            </div>
            <div
                v-if="chats.length === 0"
                class="empty-users"
            >
              <el-empty description="暂无聊天" :image-size="80" />
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
                <!-- 消息内容：根据消息类型显示 -->
                <div v-if="message.messageType === 1" class="message-text" v-html="formatMessageContent(message.content)"></div>
                <div v-else-if="message.messageType === 2" class="message-image">
                  <el-image
                    :src="message.content"
                    fit="cover"
                    style="max-width: 300px; max-height: 300px; border-radius: 4px;"
                    :preview-src-list="[message.content]"
                    preview-teleported
                  >
                    <template #error>
                      <div class="image-error">图片加载失败</div>
                    </template>
                  </el-image>
                </div>
                <div v-else class="message-text">{{ message.content }}</div>
              </div>
            </div>
            <div v-if="messages.length === 0" class="empty-message">
              <el-empty description="暂无消息，开始聊天吧~" />
            </div>
          </div>

          <!-- 输入区域 -->
          <div class="message-input">
            <!-- Emoji 选择器 -->
            <div class="emoji-picker-wrapper">
              <el-popover
                placement="top-start"
                :width="300"
                trigger="click"
                popper-class="emoji-picker-popover"
              >
                <template #reference>
                  <el-button text circle size="small" class="emoji-btn">
                    <el-icon :size="20"><ChatDotRound /></el-icon>
                  </el-button>
                </template>
                <div class="emoji-picker-container">
                  <!-- Emoji 分页显示 -->
                  <div 
                    class="emoji-picker"
                    :style="{ gridTemplateColumns: `repeat(${emojiColsPerRow}, 1fr)` }"
                  >
                    <div
                      v-for="emoji in currentPageEmojis"
                      :key="emoji"
                      class="emoji-item"
                      @click="insertEmoji(emoji)"
                    >
                      {{ emoji }}
                    </div>
                  </div>
                  <!-- 分页控制器 -->
                  <div v-if="totalEmojiPages > 1" class="emoji-pagination">
                    <el-button
                      text
                      size="small"
                      :disabled="currentEmojiPage === 0"
                      @click="currentEmojiPage--"
                    >
                      上一页
                    </el-button>
                    <span class="emoji-page-info">{{ currentEmojiPage + 1 }} / {{ totalEmojiPages }}</span>
                    <el-button
                      text
                      size="small"
                      :disabled="currentEmojiPage >= totalEmojiPages - 1"
                      @click="currentEmojiPage++"
                    >
                      下一页
                    </el-button>
                  </div>
                </div>
              </el-popover>
              <!-- 图片上传按钮 -->
              <el-upload
                :action="uploadUrl"
                :headers="uploadHeaders"
                :on-success="handleImageUploadSuccess"
                :on-error="handleImageUploadError"
                :before-upload="beforeImageUpload"
                :show-file-list="false"
                accept="image/*"
              >
                <el-button text circle size="small" class="image-btn">
                  <el-icon :size="20"><Picture /></el-icon>
                </el-button>
              </el-upload>
            </div>
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
                <span v-if="selectedChat">{{ selectedChat.name }}</span>
                <span v-else>请选择聊天</span>
              </div>
              <el-button
                  type="primary"
                  :disabled="(!inputMessage.trim() && !pendingImageUrl) || !wsConnected"
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
import { ChatDotRound, Picture } from '@element-plus/icons-vue';
import { useRoute } from 'vue-router';
import { useUserStore } from '@/stores/user';
import { useWebSocketStore, MessageType } from '@/stores/websocket';
import {
  chatMessageList,
  chatMessageSend,
  chatList,
  dictGet,
  fileUpload
} from '@/api/generated/admin';
import type {
  ChatMessageItem,
  ChatMessageListReq,
  ChatMessageSendReq,
  ChatItem,
  FileUploadResp
} from '@/api/generated/admin';
import { buildFileUrlFromResponse } from '@/utils/file';

const route = useRoute();
const userStore = useUserStore();
const wsStore = useWebSocketStore();
const currentUserId = computed(() => userStore.profile?.id || 0);
const currentUsername = computed(() => userStore.profile?.username || '');

// WebSocket 连接状态（使用全局 store）
const wsConnected = computed(() => wsStore.connected);

// 聊天相关
const selectedChatId = ref<number | null>(null);
const selectedChat = ref<ChatItem | null>(null);
const inputMessage = ref('');
const messages = ref<ChatMessageItem[]>([]);
const chats = ref<ChatItem[]>([]);
const messageListRef = ref<HTMLElement>();

// Emoji 列表
const emojiList = [
  '😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂', '🙂', '🙃',
  '😉', '😊', '😇', '🥰', '😍', '🤩', '😘', '😗', '😚', '😙',
  '😋', '😛', '😜', '🤪', '😝', '🤑', '🤗', '🤭', '🤫', '🤔',
  '🤐', '🤨', '😐', '😑', '😶', '😏', '😒', '🙄', '😬', '🤥',
  '😌', '😔', '😪', '🤤', '😴', '😷', '🤒', '🤕', '🤢', '🤮',
  '👍', '👎', '👌', '✌️', '🤞', '🤟', '🤘', '👏', '🙌', '👐',
  '❤️', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❣️'
];

// 计算每页显示的emoji数量
const emojisPerPage = computed(() => emojiColsPerRow.value * emojiRows.value);

// 计算总页数
const totalEmojiPages = computed(() => Math.ceil(emojiList.length / emojisPerPage.value));

// 当前页显示的emoji列表
const currentPageEmojis = computed(() => {
  const start = currentEmojiPage.value * emojisPerPage.value;
  const end = start + emojisPerPage.value;
  return emojiList.slice(start, end);
});

// 图片上传相关
const pendingImageUrl = ref<string>('');
const baseUrl = computed(() => import.meta.env.VITE_API_BASE_URL || '');
const uploadUrl = computed(() => `${baseUrl.value}/api/v1/files/upload`);
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${userStore.token}`
}));

// 格式化聊天描述：部门-角色-用户昵称（仅私聊）
const formatChatDesc = (chat: ChatItem): string => {
  if (chat.chatType !== 1) {
    return chat.description || '';
  }
  const parts: string[] = [];
  if (chat.departmentName) {
    parts.push(chat.departmentName);
  }
  if (chat.roleNames && chat.roleNames.length > 0) {
    parts.push(chat.roleNames.join('、'));
  }
  if (chat.nickname) {
    parts.push(chat.nickname);
  }
  return parts.join('-') || chat.username || '';
};

// 聊天配置：消息数量限制（从字典获取，默认30）
const chatMessageLimit = ref(30);
// Emoji分页配置（从字典获取）
const emojiColsPerRow = ref(8); // 每行显示数量（x），默认8
const emojiRows = ref(3); // 显示行数（y），默认3
const currentEmojiPage = ref(0); // 当前页码

// 查询参数
const query = reactive<ChatMessageListReq>({
  page: 1,
  pageSize: 30, // 默认30，将从字典加载后更新
  chatId: 0
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
      } else if (newMessage.type === 'leave') {
        ElMessage.info(`${newMessage.fromName} 离开了聊天室`);
      }
    }
);

// 处理聊天消息
const handleChatMessage = (data: any) => {
  // 检查是否是当前选中的聊天
  if (selectedChatId.value && data.chatId && Number(data.chatId) === Number(selectedChatId.value)) {
    // 收到新消息，添加到消息列表
    const newMessage: ChatMessageItem = {
      id: data.messageId || Date.now(),
      chatId: data.chatId || 0,
      fromUserId: data.fromId,
      fromUserName: data.fromName,
      content: data.content,
      messageType: data.messageType || 1, // 支持图片消息类型
      status: 1,
      createdAt: data.createdAt || Math.floor(Date.now() / 1000) // 秒级时间戳
    };
    messages.value.push(newMessage);
    // 如果消息数量超过限制，只保留最新的N条
    if (messages.value.length > chatMessageLimit.value) {
      messages.value = messages.value.slice(-chatMessageLimit.value);
    }
    scrollToBottom();
  }
};

// 插入 Emoji
const insertEmoji = (emoji: string) => {
  const textarea = document.querySelector('.message-input textarea') as HTMLTextAreaElement;
  if (textarea) {
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    inputMessage.value = inputMessage.value.substring(0, start) + emoji + inputMessage.value.substring(end);
    // 设置光标位置
    nextTick(() => {
      textarea.focus();
      textarea.setSelectionRange(start + emoji.length, start + emoji.length);
    });
  } else {
    inputMessage.value += emoji;
  }
};

// 格式化消息内容（用于显示 Emoji）
const formatMessageContent = (content: string) => {
  // 转义 HTML 特殊字符，但保留 Emoji
  return content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
    .replace(/\n/g, '<br>');
};

// 图片上传前验证
const beforeImageUpload = (file: File) => {
  const isImage = file.type.startsWith('image/');
  if (!isImage) {
    ElMessage.error('只能上传图片文件！');
    return false;
  }
  const isLt5M = file.size / 1024 / 1024 < 5;
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB！');
    return false;
  }
  return true;
};

// 图片上传成功
const handleImageUploadSuccess = async (response: FileUploadResp) => {
  try {
    const fullUrl = buildFileUrlFromResponse(response);
    pendingImageUrl.value = fullUrl;
    ElMessage.success('图片上传成功，点击发送按钮发送');
  } catch (err: any) {
    ElMessage.error('图片上传失败：' + (err.message || '未知错误'));
  }
};

// 图片上传失败
const handleImageUploadError = (error: any) => {
  ElMessage.error('图片上传失败：' + (error.message || '未知错误'));
};

// 发送消息
const handleSendMessage = async () => {
  // 检查是否有文本消息或图片
  const hasText = inputMessage.value.trim();
  const hasImage = pendingImageUrl.value;

  if (!hasText && !hasImage) {
    return;
  }

  if (!wsConnected.value) {
    ElMessage.warning('WebSocket 未连接，请先连接');
    return;
  }

  if (!selectedChatId.value) {
    ElMessage.warning('请先选择一个聊天');
    return;
  }

  try {
    // 如果有图片，发送图片消息
    if (hasImage) {
      const req: ChatMessageSendReq = {
        chatId: selectedChatId.value,
        content: pendingImageUrl.value,
        messageType: 2 // 图片消息
      };
      await chatMessageSend(req);
      pendingImageUrl.value = ''; // 清空待发送的图片
    }

    // 如果有文本，发送文本消息
    if (hasText) {
      const messageContent = inputMessage.value.trim();
      inputMessage.value = '';
      
      const req: ChatMessageSendReq = {
        chatId: selectedChatId.value,
        content: messageContent,
        messageType: 1 // 文本消息
      };
      await chatMessageSend(req);
    }

    // 消息会通过 WebSocket 推送回来，不需要手动添加到列表
  } catch (err: any) {
    ElMessage.error(err.message || '发送消息失败');
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
  if (!selectedChatId.value) {
    messages.value = [];
    return;
  }

  try {
    // 重置查询参数
    query.page = 1;
    query.pageSize = chatMessageLimit.value; // 使用从字典获取的限制值
    query.chatId = selectedChatId.value;

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

// 加载聊天列表
const loadChats = async () => {
  try {
    const resp = await chatList();
    chats.value = resp.list || [];
    // 如果没有选中的聊天，默认选中第一个
    if (chats.value.length > 0 && !selectedChatId.value) {
      selectChat(chats.value[0]);
    }
  } catch (err: any) {
    console.error('加载聊天列表失败:', err);
  }
};

// 选择聊天
const selectChat = (chat: ChatItem) => {
  selectedChatId.value = chat.chatId;
  selectedChat.value = chat;
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

  // 先加载聊天配置，再加载聊天列表
  await loadChatConfig();
  await loadChats();
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

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.user-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
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

.message-text {
  padding: 10px 14px;
  border-radius: 8px;
  background-color: var(--el-bg-color);
  font-size: 14px;
  line-height: 1.5;
  word-wrap: break-word;
  white-space: pre-wrap;
}

.message-image {
  max-width: 300px;
  margin-top: 4px;

.image-error {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 200px;
  height: 200px;
  background-color: var(--el-fill-color-light);
  color: var(--el-text-color-placeholder);
  border-radius: 4px;
}
}

.message-input {
  padding: 15px;
  border-top: 1px solid var(--el-border-color-light);
  background-color: var(--el-bg-color);

.emoji-picker-wrapper {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;

.emoji-btn,
.image-btn {
  color: var(--el-text-color-regular);
  
  &:hover {
    color: var(--el-color-primary);
  }
}
}

.emoji-picker-container {
  display: flex;
  flex-direction: column;
}

.emoji-picker {
  display: grid;
  gap: 4px;
  padding: 8px;
  min-width: 200px;

.emoji-item {
  font-size: 20px;
  padding: 4px;
  cursor: pointer;
  text-align: center;
  border-radius: 4px;
  transition: background-color 0.2s;

&:hover {
   background-color: var(--el-fill-color-light);
 }
}
}

.emoji-pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  border-top: 1px solid var(--el-border-color-lighter);
  font-size: 12px;
}

.emoji-page-info {
  color: var(--el-text-color-secondary);
}

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
