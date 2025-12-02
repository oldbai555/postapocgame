package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
)

// GetFriendSys 获取好友系统适配器
func GetFriendSys(ctx context.Context) *FriendSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysFriend))
	if system == nil {
		return nil
	}
	friendSys, ok := system.(*FriendSystemAdapter)
	if !ok || !friendSys.IsOpened() {
		return nil
	}
	return friendSys
}

// Ensure interface compatibility
var _ iface.ISystem = (*FriendSystemAdapter)(nil)
