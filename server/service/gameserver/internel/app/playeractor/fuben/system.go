package fuben

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/app/playeractor/sysbase"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// FuBenSystemAdapter 副本系统适配器
type FuBenSystemAdapter struct {
	*sysbase.BaseSystem
	deps                    Deps
	enterDungeonUseCase     *EnterDungeonUseCase
	getDungeonRecordUseCase *GetDungeonRecordUseCase
}

// NewFuBenSystemAdapter 创建副本系统适配器
func NewFuBenSystemAdapter(rt *runtime.Runtime) *FuBenSystemAdapter {
	d := depsFromRuntime(rt)
	return &FuBenSystemAdapter{
		BaseSystem:              sysbase.NewBaseSystem(uint32(protocol.SystemId_SysFuBen)),
		deps:                    d,
		enterDungeonUseCase:     NewEnterDungeonUseCase(d),
		getDungeonRecordUseCase: NewGetDungeonRecordUseCase(d),
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
	dungeonData, err := a.deps.PlayerRepo.GetDungeonData(ctx)
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

// 注册系统工厂
// RegisterSystemFactory 注册副本系统工厂（由 register.RegisterAll 调用）
func RegisterSystemFactory(rt *runtime.Runtime) {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysFuBen), func() iface.ISystem {
		return NewFuBenSystemAdapter(rt)
	})

	// 协议注册由 handler.go 的 init 负责
}
