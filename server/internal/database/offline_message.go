package database

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
)

// OfflineMessage 离线消息表
type OfflineMessage struct {
	ID         uint   `gorm:"primaryKey"`
	TargetId   uint64 `gorm:"not null;index"` // 目标角色ID
	ChatType   uint32 `gorm:"not null"`       // 聊天类型
	SenderId   uint64 `gorm:"not null"`       // 发送者ID
	SenderName string `gorm:"size:64"`        // 发送者名称
	Content    string `gorm:"type:text"`      // 消息内容
	Timestamp  int64  `gorm:"not null;index"` // 时间戳（毫秒）
	CreatedAt  int64  `gorm:"autoCreateTime"`
}

// SaveOfflineMessage 保存离线消息
func SaveOfflineMessage(targetId uint64, chatMsg *protocol.ChatMessage) error {
	msg := &OfflineMessage{
		TargetId:   targetId,
		ChatType:   uint32(chatMsg.ChatType),
		SenderId:   chatMsg.SenderId,
		SenderName: chatMsg.SenderName,
		Content:    chatMsg.Content,
		Timestamp:  chatMsg.Timestamp,
	}
	return DB.Create(msg).Error
}

// GetOfflineMessages 获取离线消息（最多20条，按时间倒序）
func GetOfflineMessages(targetId uint64) ([]*protocol.ChatMessage, error) {
	var messages []OfflineMessage
	result := DB.Where("target_id = ?", targetId).
		Order("timestamp DESC").
		Limit(20).
		Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}

	chatMessages := make([]*protocol.ChatMessage, 0, len(messages))
	// 反转顺序，使最旧的消息在前
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		chatMessages = append(chatMessages, &protocol.ChatMessage{
			ChatType:   protocol.ChatType(msg.ChatType),
			SenderId:   msg.SenderId,
			SenderName: msg.SenderName,
			TargetId:   targetId,
			Content:    msg.Content,
			Timestamp:  msg.Timestamp,
		})
	}

	return chatMessages, nil
}

// DeleteOfflineMessages 删除离线消息
func DeleteOfflineMessages(targetId uint64) error {
	return DB.Where("target_id = ?", targetId).Delete(&OfflineMessage{}).Error
}

// CleanExpiredOfflineMessages 清理过期的离线消息（超过7天）
func CleanExpiredOfflineMessages() error {
	expireTime := servertime.Now().AddDate(0, 0, -7).UnixMilli()
	return DB.Where("timestamp < ?", expireTime).Delete(&OfflineMessage{}).Error
}

// GetOfflineMessageCount 获取离线消息数量
func GetOfflineMessageCount(targetId uint64) (int64, error) {
	var count int64
	result := DB.Model(&OfflineMessage{}).Where("target_id = ?", targetId).Count(&count)
	return count, result.Error
}

// DeleteOldOfflineMessages 删除超出限制的旧消息（保留最新的20条）
func DeleteOldOfflineMessages(targetId uint64) error {
	// 获取消息总数
	var count int64
	if err := DB.Model(&OfflineMessage{}).Where("target_id = ?", targetId).Count(&count).Error; err != nil {
		return err
	}

	// 如果消息数量超过20条，删除旧消息
	if count > 20 {
		// 获取第20条消息的ID（按时间戳倒序）
		var messages []OfflineMessage
		result := DB.Where("target_id = ?", targetId).
			Order("timestamp DESC").
			Limit(20).
			Find(&messages)
		if result.Error != nil {
			return result.Error
		}

		if len(messages) > 0 {
			// 获取第20条消息的时间戳
			oldestTimestamp := messages[len(messages)-1].Timestamp
			// 删除时间戳小于该值的消息（保留最新的20条）
			return DB.Where("target_id = ? AND timestamp < ?", targetId, oldestTimestamp).
				Delete(&OfflineMessage{}).Error
		}
	}

	return nil
}
