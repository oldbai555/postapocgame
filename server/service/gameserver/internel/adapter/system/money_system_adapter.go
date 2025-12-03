package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/usecase/money"
)

// MoneySystemAdapter 货币系统适配器
//
// 生命周期职责：
// - OnInit: 调用 InitMoneyDataUseCase 初始化货币数据结构和默认金币
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（加钱、扣钱、余额校验）均在 UseCase 层实现
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type MoneySystemAdapter struct {
	*BaseSystemAdapter
	addMoneyUseCase        *money.AddMoneyUseCase
	consumeMoneyUseCase    *money.ConsumeMoneyUseCase
	updateBalanceTxUseCase *money.UpdateBalanceTxUseCase
	initMoneyDataUseCase   *money.InitMoneyDataUseCase
	moneyUseCaseImpl       interfaces.MoneyUseCase
}

// NewMoneySystemAdapter 创建货币系统适配器
func NewMoneySystemAdapter() *MoneySystemAdapter {
	container := di.GetContainer()
	moneyUseCaseImpl := money.NewMoneyUseCaseImpl(container.PlayerGateway())
	addMoneyUC := money.NewAddMoneyUseCase(container.PlayerGateway(), container.EventPublisher())
	consumeUC := money.NewConsumeMoneyUseCase(container.PlayerGateway(), container.EventPublisher())
	updateBalanceTxUC := money.NewUpdateBalanceTxUseCase(container.PlayerGateway())
	initMoneyDataUC := money.NewInitMoneyDataUseCase(container.PlayerGateway())
	return &MoneySystemAdapter{
		BaseSystemAdapter:      NewBaseSystemAdapter(uint32(protocol.SystemId_SysMoney)),
		addMoneyUseCase:        addMoneyUC,
		consumeMoneyUseCase:    consumeUC,
		updateBalanceTxUseCase: updateBalanceTxUC,
		initMoneyDataUseCase:   initMoneyDataUC,
		moneyUseCaseImpl:       moneyUseCaseImpl,
	}
}

// OnInit 系统初始化
func (a *MoneySystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
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
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.addMoneyUseCase.Execute(ctx, roleID, moneyID, amount)
}

// SubMoney 扣除货币（对外接口，供其他系统调用）
func (a *MoneySystemAdapter) SubMoney(ctx context.Context, moneyID uint32, amount int64) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
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
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return 0, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return 0, err
	}
	if binaryData.MoneyData == nil || binaryData.MoneyData.MoneyMap == nil {
		return 0, nil
	}
	return binaryData.MoneyData.MoneyMap[moneyID], nil
}

// GetMoneyData 获取货币数据（用于协议）
func (a *MoneySystemAdapter) GetMoneyData(ctx context.Context) (*protocol.SiMoneyData, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.MoneyData == nil {
		return &protocol.SiMoneyData{
			MoneyMap: make(map[uint32]int64),
		}, nil
	}
	return binaryData.MoneyData, nil
}

// GetMoneyUseCase 获取 MoneyUseCase 实现（用于注入到其他系统）
func (a *MoneySystemAdapter) GetMoneyUseCase() interfaces.MoneyUseCase {
	return a.moneyUseCaseImpl
}

// UpdateBalanceTx 更新余额（兼容旧 MoneySys 接口，仅更新 BinaryData）
func (a *MoneySystemAdapter) UpdateBalanceTx(ctx context.Context, moneyID uint32, delta int64) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	// 将“余额调整 + 不足校验”的纯业务逻辑下沉到 Money 用例，适配层只负责获取 roleID 与调用
	return a.updateBalanceTxUseCase.Execute(ctx, roleID, moneyID, delta)
}

// UpdateBalanceOnlyMemory 仅更新内存状态（用于事务回滚后的恢复）
func (a *MoneySystemAdapter) UpdateBalanceOnlyMemory(ctx context.Context, moneyID uint32, amount int64) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil || binaryData.MoneyData == nil || binaryData.MoneyData.MoneyMap == nil {
		return
	}
	binaryData.MoneyData.MoneyMap[moneyID] = amount
}

// EnsureISystem 确保 MoneySystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*MoneySystemAdapter)(nil)
