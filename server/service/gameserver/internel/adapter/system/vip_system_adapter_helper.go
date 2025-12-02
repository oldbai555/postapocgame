package system

import (
	"context"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetVipSys 获取 VIP 系统适配器（兼容旧接口）
func GetVipSys(ctx context.Context) *VipSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	sys := playerRole.GetSystem(uint32(protocol.SystemId_SysVip))
	if sys == nil {
		return nil
	}
	vipSys, ok := sys.(*VipSystemAdapter)
	if !ok || !vipSys.IsOpened() {
		return nil
	}
	return vipSys
}
