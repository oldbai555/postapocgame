package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetEquipSys 获取装备系统（系统获取函数，保持向后兼容）
func GetEquipSys(ctx context.Context) *EquipSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysEquip))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysEquip)
		return nil
	}
	sys, ok := system.(*EquipSystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysEquip)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysEquip)
		return nil
	}
	return sys
}
