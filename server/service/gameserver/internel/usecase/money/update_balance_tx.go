package money

import (
	"context"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// UpdateBalanceTxUseCase 仅更新内存中的货币余额（兼容旧 MoneySys 的 Tx 语义）
// 约束：
// - 不触发事件；
// - 不直接落库，由 PlayerActor 的定期存盘统一负责；
// - 只做余额计算与不足校验，具体“何时调用”由 SystemAdapter 控制。
type UpdateBalanceTxUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewUpdateBalanceTxUseCase 创建仅更新 BinaryData 的余额调整用例
func NewUpdateBalanceTxUseCase(playerRepo repository.PlayerRepository) *UpdateBalanceTxUseCase {
	return &UpdateBalanceTxUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 在内存中调整指定货币的余额
func (uc *UpdateBalanceTxUseCase) Execute(ctx context.Context, roleID uint64, moneyID uint32, delta int64) error {
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
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
		// 余额不足视为业务失败，由调用方根据错误码选择回滚策略
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money not enough")
	}
	binaryData.MoneyData.MoneyMap[moneyID] = newAmount
	return nil
}
