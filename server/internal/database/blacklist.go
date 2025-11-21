package database

import "gorm.io/gorm"

// Blacklist 黑名单表
type Blacklist struct {
	ID          uint   `gorm:"primaryKey"`
	RoleId      uint64 `gorm:"not null;index"` // 被拉黑的角色ID
	BlockedById uint64 `gorm:"not null;index"` // 拉黑者ID
	Reason      string `gorm:"type:text"`      // 拉黑原因
	CreatedAt   int64  `gorm:"autoCreateTime"`
	UpdatedAt   int64  `gorm:"autoUpdateTime"`
}

// AddToBlacklist 添加到黑名单
func AddToBlacklist(roleId uint64, blockedById uint64, reason string) error {
	blacklist := &Blacklist{
		RoleId:      roleId,
		BlockedById: blockedById,
		Reason:      reason,
	}

	// 检查是否已存在
	var existing Blacklist
	result := DB.Where("role_id = ? AND blocked_by_id = ?", roleId, blockedById).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新记录
		return DB.Create(blacklist).Error
	} else if result.Error != nil {
		return result.Error
	} else {
		// 更新现有记录
		return DB.Model(&existing).Updates(blacklist).Error
	}
}

// RemoveFromBlacklist 从黑名单移除
func RemoveFromBlacklist(roleId uint64, blockedById uint64) error {
	return DB.Where("role_id = ? AND blocked_by_id = ?", roleId, blockedById).Delete(&Blacklist{}).Error
}

// IsInBlacklist 检查是否在黑名单中
func IsInBlacklist(roleId uint64, blockedById uint64) (bool, error) {
	var count int64
	result := DB.Model(&Blacklist{}).Where("role_id = ? AND blocked_by_id = ?", roleId, blockedById).Count(&count)
	return count > 0, result.Error
}

// GetBlacklist 获取黑名单列表
func GetBlacklist(blockedById uint64) ([]*Blacklist, error) {
	var blacklists []Blacklist
	result := DB.Where("blocked_by_id = ?", blockedById).Find(&blacklists)
	if result.Error != nil {
		return nil, result.Error
	}

	resultList := make([]*Blacklist, 0, len(blacklists))
	for i := range blacklists {
		resultList = append(resultList, &blacklists[i])
	}
	return resultList, nil
}
