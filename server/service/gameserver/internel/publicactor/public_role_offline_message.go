package publicactor

import (
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
)

// 离线消息相关逻辑

// AddOfflineMessage 添加离线消息（持久化到数据库，限制最多20条）
func (pr *PublicRole) AddOfflineMessage(targetId uint64, chatMsg *protocol.ChatMessage) {
	// 保存到数据库
	if err := database.SaveOfflineMessage(targetId, chatMsg); err != nil {
		log.Errorf("Failed to save offline message: %v", err)
		return
	}

	// 删除超出限制的旧消息（保留最新的20条）
	if err := database.DeleteOldOfflineMessages(targetId); err != nil {
		log.Warnf("Failed to delete old offline messages: %v", err)
	}

	// 更新内存缓存
	value, _ := pr.offlineMessagesMap.LoadOrStore(targetId, make([]*protocol.ChatMessage, 0))
	messages := value.([]*protocol.ChatMessage)
	messages = append(messages, chatMsg)
	// 限制最多20条
	if len(messages) > 20 {
		messages = messages[len(messages)-20:]
	}
	pr.offlineMessagesMap.Store(targetId, messages)
	log.Debugf("Added offline message for role %d, total: %d", targetId, len(messages))
}

// GetOfflineMessages 获取并清除离线消息（从数据库加载，最多20条）
func (pr *PublicRole) GetOfflineMessages(roleId uint64) []*protocol.ChatMessage {
	// 先从数据库加载
	messages, err := database.GetOfflineMessages(roleId)
	if err != nil {
		log.Errorf("Failed to load offline messages from database: %v", err)
		// 如果数据库加载失败，尝试从内存缓存获取
		value, ok := pr.offlineMessagesMap.Load(roleId)
		if ok {
			messages = value.([]*protocol.ChatMessage)
		}
	}

	// 清除内存缓存和数据库记录
	pr.offlineMessagesMap.Delete(roleId)
	if err := database.DeleteOfflineMessages(roleId); err != nil {
		log.Warnf("Failed to delete offline messages from database: %v", err)
	}

	return messages
}
