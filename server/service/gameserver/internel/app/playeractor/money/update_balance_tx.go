package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// UpdateBalanceTxUseCase 仅更新内存中的货币余额（兼容旧 MoneySys 的 Tx 语义）
// 约束：
// - 不触发事件；
// - 不直接落库，由 PlayerActor 的定期存盘统一负责；
// - 只做余额计算与不足校验，具体“何时调用”由 SystemAdapter 控制。
type UpdateBalanceTxUseCase struct {
	deps Deps
}

// NewUpdateBalanceTxUseCase 创建仅更新 BinaryData 的余额调整用例
func NewUpdateBalanceTxUseCase(deps Deps) *UpdateBalanceTxUseCase {
	return &UpdateBalanceTxUseCase{
		deps: deps,
	}
}

// Execute 在内存中调整指定货币的余额
func (uc *UpdateBalanceTxUseCase) Execute(ctx context.Context, roleID uint64, moneyID uint32, delta int64) error {
	moneyData, err := uc.deps.PlayerRepo.GetMoneyData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}
	current := moneyData.MoneyMap[moneyID]
	newAmount := current + delta
	if newAmount < 0 {
		// 余额不足视为业务失败，由调用方根据错误码选择回滚策略
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money not enough")
	}
	moneyData.MoneyMap[moneyID] = newAmount
	return nil
}
