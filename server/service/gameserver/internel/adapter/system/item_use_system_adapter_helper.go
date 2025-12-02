package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetItemUseSys 获取物品使用系统（系统获取函数，保持向后兼容）
func GetItemUseSys(ctx context.Context) *ItemUseSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysItemUse))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysItemUse)
		return nil
	}
	sys, ok := system.(*ItemUseSystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysItemUse)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysItemUse)
		return nil
	}
	return sys
}
