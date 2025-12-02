package level

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// AddExpUseCase 添加经验用例
type AddExpUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
	configManager  interfaces.ConfigManager
	// 依赖其他系统（暂时通过接口定义，后续重构时注入）
	moneyUseCase interfaces.MoneyUseCase // 可选，如果为 nil 则通过旧方式获取
	attrUseCase  interfaces.AttrUseCase  // 可选，如果为 nil 则通过旧方式获取
}

// NewAddExpUseCase 创建添加经验用例
func NewAddExpUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
	configManager interfaces.ConfigManager,
) *AddExpUseCase {
	return &AddExpUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
		configManager:  configManager,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *AddExpUseCase) SetDependencies(moneyUseCase interfaces.MoneyUseCase, attrUseCase interfaces.AttrUseCase) {
	uc.moneyUseCase = moneyUseCase
	uc.attrUseCase = attrUseCase
}

// Execute 执行添加经验用例
func (uc *AddExpUseCase) Execute(ctx context.Context, roleID uint64, exp uint64) error {
	if exp == 0 {
		return nil
	}

	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	if binaryData.LevelData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level data not initialized")
	}

	// 更新经验值
	binaryData.LevelData.Exp += int64(exp)

	// 同步更新货币系统中的经验值
	// 注意：如果 MoneyUseCase 已注入，使用新方式；否则通过旧方式（向后兼容）
	if uc.moneyUseCase != nil {
		// 新方式：通过接口调用
		if err := uc.moneyUseCase.UpdateExp(ctx, roleID, binaryData.LevelData.Exp); err != nil {
			return err
		}
	} else {
		// 旧方式：直接更新 BinaryData（向后兼容，等 MoneySys 重构后移除）
		if binaryData.MoneyData != nil {
			if binaryData.MoneyData.MoneyMap == nil {
				binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
			}
			expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
			binaryData.MoneyData.MoneyMap[expMoneyID] = binaryData.LevelData.Exp
		}
	}

	// 发布经验变化事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnPlayerExpChange, map[string]interface{}{
		"exp": binaryData.LevelData.Exp,
	})

	// 检查是否升级
	levelUpUC := NewLevelUpUseCase(uc.playerRepo, uc.eventPublisher, uc.configManager)
	levelUpUC.SetDependencies(uc.moneyUseCase, uc.attrUseCase)
	if err := levelUpUC.Execute(ctx, roleID); err != nil {
		return err
	}

	return nil
}
