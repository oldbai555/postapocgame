package database

import (
	"gorm.io/gorm"
	"postapocgame/server/internal/servertime"
)

// ServerInfo 记录平台与区服的唯一开服时间
type ServerInfo struct {
	ID               uint   `gorm:"primaryKey"`
	PlatformID       uint32 `gorm:"not null;uniqueIndex:idx_platform_srv"`
	ServerID         uint32 `gorm:"not null;uniqueIndex:idx_platform_srv"`
	ServerOpenTimeAt int64  `gorm:"not null"` // 秒级时间戳
}

// EnsureServerInfo 如果不存在则插入一条记录，并返回当前 ServerInfo
func EnsureServerInfo(platformID, serverID uint32) (*ServerInfo, error) {
	info, err := GetServerInfo(platformID, serverID)
	if err == nil {
		return info, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	newInfo := &ServerInfo{
		PlatformID:       platformID,
		ServerID:         serverID,
		ServerOpenTimeAt: servertime.Now().Unix(),
	}
	if err := DB.Create(newInfo).Error; err != nil {
		return nil, err
	}
	return newInfo, nil
}

// GetServerInfo 查询指定平台/区服的 ServerInfo
func GetServerInfo(platformID, serverID uint32) (*ServerInfo, error) {
	var info ServerInfo
	result := DB.Where("platform_id = ? AND server_id = ?", platformID, serverID).First(&info)
	if result.Error != nil {
		return nil, result.Error
	}
	return &info, nil
}
