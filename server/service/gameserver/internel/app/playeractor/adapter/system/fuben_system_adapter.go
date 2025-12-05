package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/consume"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/fuben"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// FubenSystemAdapter 副本系统适配器
//
// 生命周期职责：
// - 当前无额外初始化需求
//
// 业务逻辑：所有业务逻辑（进入副本、副本结算、记录查找）均在 UseCase 层实现
// 外部交互：通过 DungeonServerGateway 进行副本进入/结算 RPC 调用
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type FubenSystemAdapter struct {
	*BaseSystemAdapter
	enterDungeonUseCase     *fuben.EnterDungeonUseCase
	getDungeonRecordUseCase *fuben.GetDungeonRecordUseCase
}

// NewFubenSystemAdapter 创建副本系统适配器
func NewFubenSystemAdapter() *FubenSystemAdapter {
	enterDungeonUC := fuben.NewEnterDungeonUseCase(deps.PlayerGateway(), deps.ConfigGateway(), deps.DungeonServerGateway())
	getDungeonRecordUC := fuben.NewGetDungeonRecordUseCase(deps.PlayerGateway())

	consumeUseCase := consume.NewConsumeUseCase(deps.PlayerGateway(), deps.EventPublisher())
	enterDungeonUC.SetDependencies(consumeUseCase)

	return &FubenSystemAdapter{
		BaseSystemAdapter:       NewBaseSystemAdapter(uint32(protocol.SystemId_SysFuBen)),
		enterDungeonUseCase:     enterDungeonUC,
		getDungeonRecordUseCase: getDungeonRecordUC,
	}
}

// GetDungeonRecord 获取副本记录（对外接口，供其他系统调用）
func (a *FubenSystemAdapter) GetDungeonRecord(ctx context.Context, dungeonID uint32, difficulty uint32) (*protocol.DungeonRecord, error) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// 使用 UseCase 查找副本记录（纯业务逻辑已下沉）
	return a.getDungeonRecordUseCase.Execute(ctx, roleID, dungeonID, difficulty)
}

// GetDungeonData 获取副本数据（用于协议）
func (a *FubenSystemAdapter) GetDungeonData(ctx context.Context) (*protocol.SiDungeonData, error) {
	dungeonData, err := deps.PlayerGateway().GetDungeonData(ctx)
	if err != nil {
		return nil, err
	}
	return dungeonData, nil
}

// EnsureISystem 确保 FubenSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*FubenSystemAdapter)(nil)

// GetFubenSys 获取副本系统
func GetFubenSys(ctx context.Context) *FubenSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysFuBen))
	if system == nil {
		return nil
	}
	sys, ok := system.(*FubenSystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysFuBen), func() iface.ISystem {
		return NewFubenSystemAdapter()
	})

	// 协议注册由 controller 包负责
}
