package guild

import "postapocgame/server/internal/protocol"

// EnsureGuildData 确保 GuildData 初始化
func EnsureGuildData(binaryData *protocol.PlayerRoleBinaryData) *protocol.SiGuildData {
	if binaryData == nil {
		return nil
	}
	if binaryData.GuildData == nil {
		binaryData.GuildData = &protocol.SiGuildData{
			GuildId:  0,
			Position: 0,
			JoinTime: 0,
		}
	}
	return binaryData.GuildData
}
