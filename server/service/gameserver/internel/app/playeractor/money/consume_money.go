package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// ConsumeMoneyUseCase 扣除货币用例（过程化：持有 Deps，依赖显式传入）
type ConsumeMoneyUseCase struct {
	deps Deps
	// 依赖其他系统（可选）
	levelUseCase iface.MoneyUseCase // 用于处理经验
}

// NewConsumeMoneyUseCase 创建扣除货币用例
func NewConsumeMoneyUseCase(deps Deps) *ConsumeMoneyUseCase {
	return &ConsumeMoneyUseCase{
		deps: deps,
	}
}

// SetLevelUseCase 注入经验处理用例（可选）。
func (uc *ConsumeMoneyUseCase) SetLevelUseCase(levelUseCase iface.MoneyUseCase) {
	uc.levelUseCase = levelUseCase
}

// Execute 执行扣除货币用例
func (uc *ConsumeMoneyUseCase) Execute(ctx context.Context, roleID uint64, moneyID uint32, amount int64) error {
	if amount <= 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "amount must be positive")
	}

	// 特殊货币由特殊系统处理
	switch moneyID {
	case uint32(protocol.MoneyType_MoneyTypeExp):
		// 经验通常不能扣除，这里返回错误
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "experience cannot be deducted")
	default:
		// 普通货币由货币系统处理
		return uc.updateBalance(ctx, roleID, moneyID, -amount)
	}
}

// updateBalance 更新货币余额
func (uc *ConsumeMoneyUseCase) updateBalance(ctx context.Context, roleID uint64, moneyID uint32, delta int64) error {
	moneyData, err := uc.deps.PlayerRepo.GetMoneyData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	current := moneyData.MoneyMap[moneyID]
	newAmount := current + delta

	if newAmount < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "money not enough")
	}

	// 更新余额
	moneyData.MoneyMap[moneyID] = newAmount

	// 发布事件
	uc.deps.EventPublisher.PublishPlayerEvent(ctx, gevent.OnMoneyChange, map[string]interface{}{
		"money_id": moneyID,
		"amount":   newAmount,
	})

	return nil
}
