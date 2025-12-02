package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// MoneyUseCaseImpl 实现 MoneyUseCase 接口（用于 LevelSys 依赖）
// 注意：此实现是 MoneySys 的一部分，通过 System Adapter 暴露给其他系统
type MoneyUseCaseImpl struct {
	playerRepo repository.PlayerRepository
}

// NewMoneyUseCaseImpl 创建 MoneyUseCase 实现
func NewMoneyUseCaseImpl(playerRepo repository.PlayerRepository) interfaces.MoneyUseCase {
	return &MoneyUseCaseImpl{
		playerRepo: playerRepo,
	}
}

// 确保实现接口
var _ interfaces.MoneyUseCase = (*MoneyUseCaseImpl)(nil)

// UpdateExp 更新经验值（经验作为货币的一种）
func (uc *MoneyUseCaseImpl) UpdateExp(ctx context.Context, roleID uint64, exp int64) error {
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

	expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
	binaryData.MoneyData.MoneyMap[expMoneyID] = exp

	return nil
}

// GetExp 获取经验值
func (uc *MoneyUseCaseImpl) GetExp(ctx context.Context, roleID uint64) (int64, error) {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return 0, err
	}

	if binaryData.MoneyData == nil || binaryData.MoneyData.MoneyMap == nil {
		return 0, nil
	}

	expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
	return binaryData.MoneyData.MoneyMap[expMoneyID], nil
}
