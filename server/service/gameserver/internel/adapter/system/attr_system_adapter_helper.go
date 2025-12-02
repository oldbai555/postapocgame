package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetAttrSys 获取属性系统（系统获取函数，保持向后兼容）
func GetAttrSys(ctx context.Context) *AttrSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysAttr))
	if system == nil {
		return nil
	}
	attrSys, ok := system.(*AttrSystemAdapter)
	if !ok || !attrSys.IsOpened() {
		return nil
	}
	return attrSys
}
