package level

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/domain/repository"
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
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 如果level_data不存在，则初始化
	if binaryData.LevelData == nil {
		binaryData.LevelData = &protocol.SiLevelData{
			Level: 1,
			Exp:   0,
		}
	}

	// 确保等级至少为1
	if binaryData.LevelData.Level < 1 {
		binaryData.LevelData.Level = 1
	}
	if binaryData.LevelData.Exp < 0 {
		binaryData.LevelData.Exp = 0
	}

	// 同步经验到货币系统（经验作为货币的一种）
	// 统一以等级系统的经验值为准，同步到货币系统
	if binaryData.MoneyData != nil {
		if binaryData.MoneyData.MoneyMap == nil {
			binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
		}
		expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
		// 统一以等级系统的经验值为准
		binaryData.MoneyData.MoneyMap[expMoneyID] = binaryData.LevelData.Exp
	}

	return nil
}
