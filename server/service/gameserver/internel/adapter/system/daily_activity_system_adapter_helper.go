package system

import (
	"context"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetDailyActivitySys 获取日常活跃系统适配器（兼容旧接口）
func GetDailyActivitySys(ctx context.Context) *DailyActivitySystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	sys := playerRole.GetSystem(uint32(protocol.SystemId_SysDailyActivity))
	if sys == nil {
		return nil
	}
	activitySys, ok := sys.(*DailyActivitySystemAdapter)
	if !ok || !activitySys.IsOpened() {
		return nil
	}
	return activitySys
}
