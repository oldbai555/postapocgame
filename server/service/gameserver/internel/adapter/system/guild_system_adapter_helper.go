package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetGuildSys 获取公会系统
func GetGuildSys(ctx context.Context) *GuildSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysGuild))
	if system == nil {
		return nil
	}
	guildSys, ok := system.(*GuildSystemAdapter)
	if !ok || !guildSys.IsOpened() {
		return nil
	}
	return guildSys
}
