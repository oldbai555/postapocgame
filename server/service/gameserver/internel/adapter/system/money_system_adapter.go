package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/dailyactivity"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/usecase/money"
	vipusecase "postapocgame/server/service/gameserver/internel/usecase/vip"
)

// MoneySystemAdapter 货币系统适配器
type MoneySystemAdapter struct {
	*BaseSystemAdapter
	addMoneyUseCase     *money.AddMoneyUseCase
	consumeMoneyUseCase *money.ConsumeMoneyUseCase
	moneyUseCaseImpl    interfaces.MoneyUseCase
}

// NewMoneySystemAdapter 创建货币系统适配器
func NewMoneySystemAdapter() *MoneySystemAdapter {
	container := di.GetContainer()
	moneyUseCaseImpl := money.NewMoneyUseCaseImpl(container.PlayerGateway())
	addMoneyUC := money.NewAddMoneyUseCase(container.PlayerGateway(), container.EventPublisher())
	// 为特殊货币注入对应用例（VIP 经验 + 活跃点）
	vipUC := vipusecase.NewVipMoneyUseCaseImpl(container.PlayerGateway(), container.ConfigGateway())
	activeUC := dailyactivity.NewPointsUseCase(container.PlayerGateway(), container.EventPublisher())
	addMoneyUC.SetDependencies(nil, vipUC, activeUC)
	consumeUC := money.NewConsumeMoneyUseCase(container.PlayerGateway(), container.EventPublisher())
	consumeUC.SetDependencies(nil, activeUC)
	return &MoneySystemAdapter{
		BaseSystemAdapter:   NewBaseSystemAdapter(uint32(protocol.SystemId_SysMoney)),
		addMoneyUseCase:     addMoneyUC,
		consumeMoneyUseCase: consumeUC,
		moneyUseCaseImpl:    moneyUseCaseImpl,
	}
}

// OnInit 系统初始化
func (a *MoneySystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("money sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("money sys OnInit get binary data err:%v", err)
		return
	}

	// 如果money_data不存在，则初始化
	if binaryData.MoneyData == nil {
		binaryData.MoneyData = &protocol.SiMoneyData{
			MoneyMap: make(map[uint32]int64),
		}
	}

	// 如果MoneyMap为空，初始化默认金币
	if len(binaryData.MoneyData.MoneyMap) == 0 {
		defaultGoldMoneyID := uint32(protocol.MoneyType_MoneyTypeGoldCoin)
		defaultGoldAmount := int64(100000)
		binaryData.MoneyData.MoneyMap[defaultGoldMoneyID] = defaultGoldAmount
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
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}
	if binaryData.MoneyData == nil {
		binaryData.MoneyData = &protocol.SiMoneyData{
			MoneyMap: make(map[uint32]int64),
		}
	}
	if binaryData.MoneyData.MoneyMap == nil {
		binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
	}
	current := binaryData.MoneyData.MoneyMap[moneyID]
	newAmount := current + delta
	if newAmount < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money not enough")
	}
	binaryData.MoneyData.MoneyMap[moneyID] = newAmount
	return nil
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
