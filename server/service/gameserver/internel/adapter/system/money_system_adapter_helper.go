package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetMoneySys 获取货币系统（系统获取函数，保持向后兼容）
func GetMoneySys(ctx context.Context) *MoneySystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysMoney))
	if system == nil {
		return nil
	}
	sys, ok := system.(*MoneySystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}
