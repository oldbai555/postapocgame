package system

import (
	"context"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetGMSys 获取 GM 系统（系统获取函数，保持向后兼容）
func GetGMSys(ctx context.Context) *GMSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysGM))
	if system == nil {
		return nil
	}
	sys, ok := system.(*GMSystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}
