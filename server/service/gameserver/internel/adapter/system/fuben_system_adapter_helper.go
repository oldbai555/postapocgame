package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetFubenSys 获取副本系统（系统获取函数，保持向后兼容）
func GetFubenSys(ctx context.Context) *FubenSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysFuBen))
	if system == nil {
		return nil
	}
	sys, ok := system.(*FubenSystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}
