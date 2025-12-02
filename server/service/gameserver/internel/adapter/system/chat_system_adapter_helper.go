package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetChatSys 获取聊天系统
func GetChatSys(ctx context.Context) *ChatSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysChat))
	if system == nil {
		return nil
	}
	chatSys, ok := system.(*ChatSystemAdapter)
	if !ok || !chatSys.IsOpened() {
		return nil
	}
	return chatSys
}
