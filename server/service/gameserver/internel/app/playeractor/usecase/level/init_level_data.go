package level

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
)

// InitLevelDataUseCase 初始化等级数据用例
// 负责等级数据的初始化、默认值修正、经验与货币系统的同步
type InitLevelDataUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewInitLevelDataUseCase 创建初始化等级数据用例
func NewInitLevelDataUseCase(
	playerRepo repository.PlayerRepository,
) *InitLevelDataUseCase {
	return &InitLevelDataUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行初始化等级数据用例
func (uc *InitLevelDataUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	levelData, err := uc.playerRepo.GetLevelData(ctx)
	if err != nil {
		return err
	}
	moneyData, err := uc.playerRepo.GetMoneyData(ctx)
	if err != nil {
		return err
	}

	// 确保等级至少为1
	if levelData.Level < 1 {
		levelData.Level = 1
	}
	if levelData.Exp < 0 {
		levelData.Exp = 0
	}

	// 同步经验到货币系统（经验作为货币的一种）
	// 统一以等级系统的经验值为准，同步到货币系统
	expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
	// 统一以等级系统的经验值为准
	moneyData.MoneyMap[expMoneyID] = levelData.Exp

	return nil
}
