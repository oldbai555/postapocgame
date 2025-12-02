package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetBagSys 获取背包系统（系统获取函数，保持向后兼容）
func GetBagSys(ctx context.Context) *BagSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysBag))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysBag)
		return nil
	}
	sys, ok := system.(*BagSystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysBag)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysBag)
		return nil
	}
	return sys
}
