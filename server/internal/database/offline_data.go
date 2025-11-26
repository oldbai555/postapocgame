package database

import (
	"gorm.io/gorm/clause"
	"postapocgame/server/internal/protocol"
)

// OfflineData 玩家离线数据表
type OfflineData struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	RoleID    uint64 `gorm:"not null;index:idx_role_type,priority:1"`
	DataType  uint32 `gorm:"not null;index:idx_role_type,priority:2"`
	Data      []byte `gorm:"type:blob;not null"`
	Version   uint32 `gorm:"not null;default:1"`
	UpdatedAt int64  `gorm:"not null;index"`
	CreatedAt int64  `gorm:"autoCreateTime"`
}

// UpsertOfflineData 写入或更新玩家离线数据
func UpsertOfflineData(record *OfflineData) error {
	if record == nil {
		return nil
	}
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "data_type"}},
		UpdateAll: true,
	}).Create(record).Error
}

// GetOfflineData 获取单个离线数据
func GetOfflineData(roleID uint64, dataType protocol.OfflineDataType) (*OfflineData, error) {
	var record OfflineData
	result := DB.Where("role_id = ? AND data_type = ?", roleID, uint32(dataType)).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// GetAllOfflineData 获取全部离线数据
func GetAllOfflineData() ([]*OfflineData, error) {
	var records []*OfflineData
	result := DB.Find(&records)
	return records, result.Error
}

// DeleteOfflineData 删除指定离线数据
func DeleteOfflineData(roleID uint64, dataType protocol.OfflineDataType) error {
	return DB.Where("role_id = ? AND data_type = ?", roleID, uint32(dataType)).Delete(&OfflineData{}).Error
}
