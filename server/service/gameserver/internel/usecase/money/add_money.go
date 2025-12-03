package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// AddMoneyUseCase 添加货币用例
type AddMoneyUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
	// 依赖其他系统（暂时通过接口定义，后续重构时注入）
	levelUseCase interfaces.MoneyUseCase // 用于处理经验
}

// NewAddMoneyUseCase 创建添加货币用例
func NewAddMoneyUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
) *AddMoneyUseCase {
	return &AddMoneyUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *AddMoneyUseCase) SetDependencies(
	levelUseCase interfaces.MoneyUseCase,
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
		// 经验由等级系统处理
		if uc.levelUseCase != nil {
			return uc.levelUseCase.UpdateExp(ctx, roleID, amount)
		}
		// 旧方式：通过 GetLevelSys 获取（向后兼容）
		// 注意：这里违反了 Clean Architecture 原则，等 LevelSys 完全重构后移除
		return uc.updateExpLegacy(ctx, roleID, amount)
	default:
		// 普通货币由货币系统处理
		return uc.updateBalance(ctx, roleID, moneyID, amount)
	}
}

// updateBalance 更新普通货币余额
func (uc *AddMoneyUseCase) updateBalance(ctx context.Context, roleID uint64, moneyID uint32, delta int64) error {
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

	// 检查是否溢出（int64最大值）
	const maxInt64 = int64(^uint64(0) >> 1)
	if newAmount > maxInt64 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "money overflow: amount exceeds maximum")
	}

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

// updateExpLegacy 更新经验值（旧方式，向后兼容）
func (uc *AddMoneyUseCase) updateExpLegacy(ctx context.Context, roleID uint64, amount int64) error {
	// 通过 Context 获取 PlayerRole，然后调用 LevelSys（向后兼容）
	// 注意：这里违反了 Clean Architecture 原则，等 LevelSys 完全重构后移除
	// 暂时直接更新 BinaryData 中的经验值（与 LevelSys 同步）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}
	if binaryData.LevelData == nil {
		binaryData.LevelData = &protocol.SiLevelData{
			Level: 1,
			Exp:   0,
		}
	}
	// 更新经验值（与 LevelSys 同步）
	binaryData.LevelData.Exp += amount
	// 同步到货币系统
	if binaryData.MoneyData == nil {
		binaryData.MoneyData = &protocol.SiMoneyData{
			MoneyMap: make(map[uint32]int64),
		}
	}
	if binaryData.MoneyData.MoneyMap == nil {
		binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
	}
	expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
	binaryData.MoneyData.MoneyMap[expMoneyID] = binaryData.LevelData.Exp
	return nil
}
