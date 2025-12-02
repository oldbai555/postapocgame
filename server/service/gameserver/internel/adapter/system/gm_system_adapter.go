package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/iface"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GMSystemAdapter 玩家级 GM 系统适配器
type GMSystemAdapter struct {
	*BaseSystemAdapter
	mgr *GMManager
}

func NewGMSystemAdapter() *GMSystemAdapter {
	return &GMSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysGM)),
		mgr:               NewGMManager(),
	}
}

func (gm *GMSystemAdapter) OnInit(context.Context) {
	gm.mgr = NewGMManager()
}

// ExecuteCommand 执行 GM 命令（对外提供的统一入口）
func (gm *GMSystemAdapter) ExecuteCommand(ctx context.Context, gmName string, args []string) (bool, string) {
	if gm.mgr == nil {
		return false, "GM系统未初始化"
	}
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("GMSystem.ExecuteCommand: get player role error: %v", err)
		return false, err.Error()
	}
	return gm.mgr.ExecuteCommand(ctx, playerRole, gmName, args)
}

// 确保实现 ISystem 接口
var _ iface.ISystem = (*GMSystemAdapter)(nil)
