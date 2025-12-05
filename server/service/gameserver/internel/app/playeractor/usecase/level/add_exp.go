package level

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// AddExpUseCase 添加经验用例
type AddExpUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces2.EventPublisher
	configManager  interfaces2.ConfigManager
	// 依赖其他系统（通过接口注入）
	moneyUseCase  interfaces2.MoneyUseCase  // 必须：用于同步经验货币
	attrUseCase   interfaces2.AttrUseCase   // 可选：标记属性变更
	rewardUseCase interfaces2.RewardUseCase // 必须：用于下发升级奖励
}

// NewAddExpUseCase 创建添加经验用例
func NewAddExpUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces2.EventPublisher,
	configManager interfaces2.ConfigManager,
) *AddExpUseCase {
	return &AddExpUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
		configManager:  configManager,
	}
}

// SetDependencies 设置依赖
func (uc *AddExpUseCase) SetDependencies(moneyUseCase interfaces2.MoneyUseCase, attrUseCase interfaces2.AttrUseCase, rewardUseCase interfaces2.RewardUseCase) {
	uc.moneyUseCase = moneyUseCase
	uc.attrUseCase = attrUseCase
	uc.rewardUseCase = rewardUseCase
}

// Execute 执行添加经验用例
func (uc *AddExpUseCase) Execute(ctx context.Context, roleID uint64, exp uint64) error {
	if exp == 0 {
		return nil
	}

	// 获取 BinaryData（共享引用）
	levelData, err := uc.playerRepo.GetLevelData(ctx)
	if err != nil {
		return err
	}

	// 更新经验值
	levelData.Exp += int64(exp)

	// 同步更新经验货币（必须提供 MoneyUseCase）
	if uc.moneyUseCase == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money use case not injected")
	}
	if err := uc.moneyUseCase.UpdateExp(ctx, roleID, levelData.Exp); err != nil {
		return err
	}

	// 发布经验变化事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnPlayerExpChange, map[string]interface{}{
		"exp": levelData.Exp,
	})

	// 检查是否升级
	levelUpUC := NewLevelUpUseCase(uc.playerRepo, uc.eventPublisher, uc.configManager)
	levelUpUC.SetDependencies(uc.moneyUseCase, uc.attrUseCase, uc.rewardUseCase)
	if err := levelUpUC.Execute(ctx, roleID); err != nil {
		return err
	}

	return nil
}
