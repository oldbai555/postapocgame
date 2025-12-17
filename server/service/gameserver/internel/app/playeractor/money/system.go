package money

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

type MoneySystemAdapter struct {
	*sysbase.BaseSystem
	deps                   Deps
	addMoneyUseCase        *AddMoneyUseCase
	consumeMoneyUseCase    *ConsumeMoneyUseCase
	updateBalanceTxUseCase *UpdateBalanceTxUseCase
	initMoneyDataUseCase   *InitMoneyDataUseCase
	moneyUseCaseImpl       iface.MoneyUseCase
}

// NewMoneySystemAdapter 创建货币系统适配器
func NewMoneySystemAdapter(rt *runtime.Runtime) *MoneySystemAdapter {
	d := depsFromRuntime(rt)
	moneyUseCaseImpl := NewMoneyUseCaseImpl(d)
	addMoneyUC := NewAddMoneyUseCase(d)
	addMoneyUC.SetLevelUseCase(moneyUseCaseImpl)
	consumeUC := NewConsumeMoneyUseCase(d)
	updateBalanceTxUC := NewUpdateBalanceTxUseCase(d)
	initMoneyDataUC := NewInitMoneyDataUseCase(d)
	return &MoneySystemAdapter{
		BaseSystem:             sysbase.NewBaseSystem(uint32(protocol.SystemId_SysMoney)),
		deps:                   d,
		addMoneyUseCase:        addMoneyUC,
		consumeMoneyUseCase:    consumeUC,
		updateBalanceTxUseCase: updateBalanceTxUC,
		initMoneyDataUseCase:   initMoneyDataUC,
		moneyUseCaseImpl:       moneyUseCaseImpl,
	}
}

// OnInit 系统初始化
func (a *MoneySystemAdapter) OnInit(ctx context.Context) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("money sys OnInit get role err:%v", err)
		return
	}
	// 初始化货币数据（包括 MoneyData 结构、默认金币注入等业务逻辑）
	if err := a.initMoneyDataUseCase.Execute(ctx, roleID); err != nil {
		log.Errorf("money sys OnInit init money data err:%v", err)
		return
	}
	log.Infof("MoneySys initialized")
}

// AddMoney 添加货币（对外接口，供其他系统调用）
func (a *MoneySystemAdapter) AddMoney(ctx context.Context, moneyID uint32, amount int64) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.addMoneyUseCase.Execute(ctx, roleID, moneyID, amount)
}

// SubMoney 扣除货币（对外接口，供其他系统调用）
func (a *MoneySystemAdapter) SubMoney(ctx context.Context, moneyID uint32, amount int64) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.consumeMoneyUseCase.Execute(ctx, roleID, moneyID, amount)
}

// CostMoney 扣除货币（兼容旧接口，调用SubMoney）
func (a *MoneySystemAdapter) CostMoney(ctx context.Context, moneyID uint32, amount int64) error {
	return a.SubMoney(ctx, moneyID, amount)
}

// GetAmount 获取货币数量
func (a *MoneySystemAdapter) GetAmount(ctx context.Context, moneyID uint32) (int64, error) {
	moneyData, err := a.deps.PlayerRepo.GetMoneyData(ctx)
	if err != nil {
		return 0, err
	}
	return moneyData.MoneyMap[moneyID], nil
}

// GetMoneyData 获取货币数据（用于协议）
func (a *MoneySystemAdapter) GetMoneyData(ctx context.Context) (*protocol.SiMoneyData, error) {
	moneyData, err := a.deps.PlayerRepo.GetMoneyData(ctx)
	if err != nil {
		return nil, err
	}
	return moneyData, nil
}

// GetMoneyUseCase 获取 MoneyUseCase 实现（用于注入到其他系统）
func (a *MoneySystemAdapter) GetMoneyUseCase() iface.MoneyUseCase {
	return a.moneyUseCaseImpl
}

// UpdateBalanceTx 更新余额（兼容旧 MoneySys 接口，仅更新 BinaryData）
func (a *MoneySystemAdapter) UpdateBalanceTx(ctx context.Context, moneyID uint32, delta int64) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	// 将“余额调整 + 不足校验”的纯业务逻辑下沉到 Money 用例，适配层只负责获取 roleID 与调用
	return a.updateBalanceTxUseCase.Execute(ctx, roleID, moneyID, delta)
}

// UpdateBalanceOnlyMemory 仅更新内存状态（用于事务回滚后的恢复）
func (a *MoneySystemAdapter) UpdateBalanceOnlyMemory(ctx context.Context, moneyID uint32, amount int64) error {
	moneyData, err := a.deps.PlayerRepo.GetMoneyData(ctx)
	if err != nil {
		return err
	}
	moneyData.MoneyMap[moneyID] = amount
	return nil
}

// EnsureISystem 确保 MoneySystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*MoneySystemAdapter)(nil)

// GetMoneySys 获取货币系统
func GetMoneySys(ctx context.Context) *MoneySystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysMoney))
	if system == nil {
		return nil
	}
	sys, ok := system.(*MoneySystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}

// 注册系统工厂和协议
// RegisterSystemFactory 注册货币系统工厂（由 register.RegisterAll 调用）
func RegisterSystemFactory(rt *runtime.Runtime) {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysMoney), func() iface.ISystem {
		return NewMoneySystemAdapter(rt)
	})

	// 协议注册由 controller 包负责
}
