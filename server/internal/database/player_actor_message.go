package database

import (
	"postapocgame/server/internal/servertime"
)

// PlayerActorMessage 玩家Actor消息表
type PlayerActorMessage struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	RoleId    uint64 `gorm:"not null;index"`     // 目标角色ID
	MsgType   int32  `gorm:"not null;index"`     // 消息类型
	MsgData   []byte `gorm:"type:blob;not null"` // 消息数据（序列化后的 proto 字节）
	CreatedAt int64  `gorm:"autoCreateTime"`     // 创建时间（秒）
}

// SavePlayerActorMessage 保存玩家消息
// 注意：此函数不检查消息数量限制，限制检查由 MessageSystemAdapter.RunOne 统一处理
func SavePlayerActorMessage(roleId uint64, msgType int32, msgData []byte) error {
	msg := &PlayerActorMessage{
		RoleId:    roleId,
		MsgType:   msgType,
		MsgData:   msgData,
		CreatedAt: servertime.Now().Unix(),
	}
	return DB.Create(msg).Error
}

// LoadPlayerActorMessages 根据角色ID加载消息，支持根据消息ID增量加载
func LoadPlayerActorMessages(roleId uint64, afterMsgId uint64) ([]*PlayerActorMessage, error) {
	var messages []*PlayerActorMessage
	query := DB.Where("role_id = ?", roleId)
	if afterMsgId > 0 {
		query = query.Where("id > ?", afterMsgId)
	}
	err := query.Order("id ASC").Find(&messages).Error
	return messages, err
}

// DeletePlayerActorMessage 删除单条消息
func DeletePlayerActorMessage(msgId uint64) error {
	return DB.Where("id = ?", msgId).Delete(&PlayerActorMessage{}).Error
}

// DeletePlayerActorMessages 删除指定角色的所有消息
func DeletePlayerActorMessages(roleId uint64) error {
	return DB.Where("role_id = ?", roleId).Delete(&PlayerActorMessage{}).Error
}

// GetPlayerActorMessageCount 获取指定角色的消息数量
func GetPlayerActorMessageCount(roleId uint64) (int64, error) {
	var count int64
	err := DB.Model(&PlayerActorMessage{}).Where("role_id = ?", roleId).Count(&count).Error
	return count, err
}

// DeleteExpiredPlayerActorMessages 删除过期消息（超过指定天数的消息）
func DeleteExpiredPlayerActorMessages(expireDays int) (int64, error) {
	if expireDays <= 0 {
		return 0, nil
	}
	expireTime := servertime.Now().Unix() - int64(expireDays*24*3600)
	result := DB.Where("created_at < ?", expireTime).Delete(&PlayerActorMessage{})
	return result.RowsAffected, result.Error
}

// DeleteOldestPlayerActorMessages 删除指定角色的最旧消息，保留最新的 maxCount 条
func DeleteOldestPlayerActorMessages(roleId uint64, maxCount int64) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	var count int64
	err := DB.Model(&PlayerActorMessage{}).Where("role_id = ?", roleId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	if count <= maxCount {
		return 0, nil
	}
	// 获取需要保留的最新消息的ID
	var keepMessages []PlayerActorMessage
	err = DB.Where("role_id = ?", roleId).Order("id DESC").Limit(int(maxCount)).Find(&keepMessages).Error
	if err != nil {
		return 0, err
	}
	if len(keepMessages) == 0 {
		return 0, nil
	}
	// 获取需要保留的最小ID
	minKeepId := keepMessages[len(keepMessages)-1].ID
	// 删除比最小ID更小的消息
	result := DB.Where("role_id = ? AND id < ?", roleId, minKeepId).Delete(&PlayerActorMessage{})
	return result.RowsAffected, result.Error
}
