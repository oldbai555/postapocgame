package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/iface"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GMSystemAdapter 玩家级 GM 系统适配器
//
// 生命周期职责：
// - OnInit: 初始化 GMManager
// - 其他生命周期: 暂未使用
//
// 业务逻辑：GM 指令解析、权限校验与实际执行逻辑逐步下沉到 GM 用例与工具函数中
// 协议注册：在 gm_system_adapter_init.go 中通过 OnSrvStart 事件注册 C2SGMCommand 协议
// 注意：主要逻辑为执行 GM 命令，属于框架层面的命令处理，保留在适配层符合 Clean Architecture 原则
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
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
