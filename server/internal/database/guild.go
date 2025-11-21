package database

import (
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"postapocgame/server/internal/protocol"
)

// Guild 公会表
type Guild struct {
	ID           uint   `gorm:"primaryKey"`
	GuildId      uint64 `gorm:"not null;uniqueIndex"`
	GuildName    string `gorm:"not null;size:64;index"`
	CreatorId    uint64 `gorm:"not null"`
	Level        uint32 `gorm:"not null;default:1"`
	CreateTime   int64  `gorm:"not null"`
	Announcement string `gorm:"type:text"`
	BinaryData   []byte `gorm:"type:blob"` // GuildData的二进制数据
	CreatedAt    int64  `gorm:"autoCreateTime"`
	UpdatedAt    int64  `gorm:"autoUpdateTime"`
}

// SaveGuild 保存公会数据
func SaveGuild(guildData *protocol.GuildData) error {
	data, err := proto.Marshal(guildData)
	if err != nil {
		return err
	}

	guild := &Guild{
		GuildId:      guildData.GuildId,
		GuildName:    guildData.GuildName,
		CreatorId:    guildData.CreatorId,
		Level:        guildData.Level,
		CreateTime:   guildData.CreateTime,
		Announcement: guildData.Announcement,
		BinaryData:   data,
	}

	// 使用GuildId作为唯一键，如果存在则更新，否则创建
	var existingGuild Guild
	result := DB.Where("guild_id = ?", guildData.GuildId).First(&existingGuild)
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新记录
		return DB.Create(guild).Error
	} else if result.Error != nil {
		return result.Error
	} else {
		// 更新现有记录
		return DB.Model(&existingGuild).Updates(guild).Error
	}
}

// GetGuild 获取公会数据
func GetGuild(guildId uint64) (*protocol.GuildData, error) {
	var guild Guild
	result := DB.Where("guild_id = ?", guildId).First(&guild)
	if result.Error != nil {
		return nil, result.Error
	}

	guildData := &protocol.GuildData{}
	if err := proto.Unmarshal(guild.BinaryData, guildData); err != nil {
		return nil, err
	}

	return guildData, nil
}

// GetAllGuilds 获取所有公会数据
func GetAllGuilds() ([]*protocol.GuildData, error) {
	var guilds []Guild
	result := DB.Find(&guilds)
	if result.Error != nil {
		return nil, result.Error
	}

	guildDataList := make([]*protocol.GuildData, 0, len(guilds))
	for _, guild := range guilds {
		guildData := &protocol.GuildData{}
		if err := proto.Unmarshal(guild.BinaryData, guildData); err != nil {
			continue
		}
		guildDataList = append(guildDataList, guildData)
	}

	return guildDataList, nil
}

// DeleteGuild 删除公会
func DeleteGuild(guildId uint64) error {
	return DB.Where("guild_id = ?", guildId).Delete(&Guild{}).Error
}

// CheckGuildNameExists 检查公会名称是否已存在
func CheckGuildNameExists(guildName string) (bool, error) {
	var count int64
	result := DB.Model(&Guild{}).Where("guild_name = ?", guildName).Count(&count)
	return count > 0, result.Error
}
