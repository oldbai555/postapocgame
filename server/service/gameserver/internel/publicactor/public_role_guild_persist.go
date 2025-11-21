package publicactor

import (
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
)

// 公会持久化相关逻辑

// SetGuild 设置公会数据（并持久化到数据库）
func (pr *PublicRole) SetGuild(guildId uint64, guild *protocol.GuildData) {
	pr.guildMap.Store(guildId, guild)
	// 持久化到数据库
	if err := database.SaveGuild(guild); err != nil {
		log.Errorf("Failed to save guild to database: %v", err)
	}
}

// DeleteGuild 删除公会（并从数据库删除）
func (pr *PublicRole) DeleteGuild(guildId uint64) {
	pr.guildMap.Delete(guildId)
	// 从数据库删除
	if err := database.DeleteGuild(guildId); err != nil {
		log.Errorf("Failed to delete guild from database: %v", err)
	}
}
