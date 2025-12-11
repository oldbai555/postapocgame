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

type FuBenSystemAdapter struct {
	*BaseSystemAdapter
	enterDungeonUseCase     *fuben.EnterDungeonUseCase
	getDungeonRecordUseCase *fuben.GetDungeonRecordUseCase
}

// NewFuBenSystemAdapter 创建副本系统适配器
func NewFuBenSystemAdapter() *FuBenSystemAdapter {
	enterDungeonUC := fuben.NewEnterDungeonUseCase(deps.PlayerGateway(), deps.ConfigGateway(), deps.DungeonServerGateway())
	getDungeonRecordUC := fuben.NewGetDungeonRecordUseCase(deps.PlayerGateway())

	consumeUseCase := consume.NewConsumeUseCase(deps.PlayerGateway(), deps.EventPublisher())
	enterDungeonUC.SetDependencies(consumeUseCase)

	return &FuBenSystemAdapter{
		BaseSystemAdapter:       NewBaseSystemAdapter(uint32(protocol.SystemId_SysFuBen)),
		enterDungeonUseCase:     enterDungeonUC,
		getDungeonRecordUseCase: getDungeonRecordUC,
	}
}

// GetDungeonRecord 获取副本记录（对外接口，供其他系统调用）
func (a *FuBenSystemAdapter) GetDungeonRecord(ctx context.Context, dungeonID uint32, difficulty uint32) (*protocol.DungeonRecord, error) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// 使用 UseCase 查找副本记录（纯业务逻辑已下沉）
	return a.getDungeonRecordUseCase.Execute(ctx, roleID, dungeonID, difficulty)
}

// GetDungeonData 获取副本数据（用于协议）
func (a *FuBenSystemAdapter) GetDungeonData(ctx context.Context) (*protocol.SiDungeonData, error) {
	dungeonData, err := deps.PlayerGateway().GetDungeonData(ctx)
	if err != nil {
		return nil, err
	}
	return dungeonData, nil
}

// EnsureISystem 确保 FuBenSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*FuBenSystemAdapter)(nil)

// GetFubenSys 获取副本系统
func GetFubenSys(ctx context.Context) *FuBenSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysFuBen))
	if system == nil {
		return nil
	}
	sys, ok := system.(*FuBenSystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysFuBen), func() iface.ISystem {
		return NewFuBenSystemAdapter()
	})

	// 协议注册由 controller 包负责
}
