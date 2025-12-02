package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// ConsumeMoneyUseCase 扣除货币用例
type ConsumeMoneyUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
	// 依赖其他系统（暂时通过接口定义，后续重构时注入）
	levelUseCase    interfaces.MoneyUseCase // 用于处理经验
	activityUseCase interfaces.MoneyUseCase // 用于处理活跃点（待实现）
}

// NewConsumeMoneyUseCase 创建扣除货币用例
func NewConsumeMoneyUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
) *ConsumeMoneyUseCase {
	return &ConsumeMoneyUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *ConsumeMoneyUseCase) SetDependencies(
	levelUseCase interfaces.MoneyUseCase,
	activityUseCase interfaces.MoneyUseCase,
) {
	uc.levelUseCase = levelUseCase
	uc.activityUseCase = activityUseCase
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
	case uint32(protocol.MoneyType_MoneyTypeActivePoint):
		// 活跃点由活跃度系统处理
		if uc.activityUseCase != nil {
			return uc.activityUseCase.UpdateExp(ctx, roleID, -amount)
		}
		// 旧方式保留为空实现
		return uc.consumeActivePointLegacy(ctx, roleID, amount)
	default:
		// 普通货币由货币系统处理
		return uc.updateBalance(ctx, roleID, moneyID, -amount)
	}
}

// updateBalance 更新货币余额
func (uc *ConsumeMoneyUseCase) updateBalance(ctx context.Context, roleID uint64, moneyID uint32, delta int64) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 确保 MoneyData 已初始化
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
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "money not enough")
	}

	// 更新余额
	binaryData.MoneyData.MoneyMap[moneyID] = newAmount

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnMoneyChange, map[string]interface{}{
		"money_id": moneyID,
		"amount":   newAmount,
	})

	return nil
}

// consumeActivePointLegacy 扣除活跃点（旧方式，向后兼容）
func (uc *ConsumeMoneyUseCase) consumeActivePointLegacy(ctx context.Context, roleID uint64, amount int64) error {
	// 暂时不处理，等 DailyActivitySys 重构后实现
	_ = ctx
	_ = roleID
	_ = amount
	return nil
}
