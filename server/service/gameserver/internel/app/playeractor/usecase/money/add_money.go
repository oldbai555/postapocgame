package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// AddMoneyUseCase 添加货币用例
type AddMoneyUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces2.EventPublisher
	// 依赖其他系统（暂时通过接口定义，后续重构时注入）
	levelUseCase interfaces2.MoneyUseCase // 用于处理经验
}

// NewAddMoneyUseCase 创建添加货币用例
func NewAddMoneyUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces2.EventPublisher,
) *AddMoneyUseCase {
	return &AddMoneyUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *AddMoneyUseCase) SetDependencies(
	levelUseCase interfaces2.MoneyUseCase,
) {
	uc.levelUseCase = levelUseCase
}

// Execute 执行添加货币用例
func (uc *AddMoneyUseCase) Execute(ctx context.Context, roleID uint64, moneyID uint32, amount int64) error {
	if amount == 0 {
		return nil
	}
	if amount < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "amount must be positive")
	}

	// 特殊货币由特殊系统处理
	switch moneyID {
	case uint32(protocol.MoneyType_MoneyTypeExp):
		// 经验由等级系统处理，必须注入 levelUseCase
		if uc.levelUseCase == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money use case not injected")
		}
		return uc.levelUseCase.UpdateExp(ctx, roleID, amount)
	default:
		// 普通货币由货币系统处理
		return uc.updateBalance(ctx, roleID, moneyID, amount)
	}
}

// updateBalance 更新普通货币余额
func (uc *AddMoneyUseCase) updateBalance(ctx context.Context, roleID uint64, moneyID uint32, delta int64) error {
	// 获取 BinaryData（共享引用）
	moneyData, err := uc.playerRepo.GetMoneyData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	current := moneyData.MoneyMap[moneyID]
	newAmount := current + delta

	// 检查是否溢出（int64最大值）
	const maxInt64 = int64(^uint64(0) >> 1)
	if newAmount > maxInt64 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "money overflow: amount exceeds maximum")
	}

	if newAmount < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "money not enough")
	}

	// 更新余额
	moneyData.MoneyMap[moneyID] = newAmount

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnMoneyChange, map[string]interface{}{
		"money_id": moneyID,
		"amount":   newAmount,
	})

	return nil
}
