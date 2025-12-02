package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetShopSys 获取商城系统（系统获取函数，保持向后兼容）
func GetShopSys(ctx context.Context) *ShopSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get ShopSys player role failed: %v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysShop))
	if system == nil {
		return nil
	}
	sys, ok := system.(*ShopSystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}
