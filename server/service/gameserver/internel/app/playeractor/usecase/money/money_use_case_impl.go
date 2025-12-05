package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
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
	moneyData, err := uc.playerRepo.GetMoneyData(ctx)
	if err != nil {
		return err
	}
	expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
	moneyData.MoneyMap[expMoneyID] = exp
	return nil
}

// GetExp 获取经验值
func (uc *MoneyUseCaseImpl) GetExp(ctx context.Context, roleID uint64) (int64, error) {
	moneyData, err := uc.playerRepo.GetMoneyData(ctx)
	if err != nil {
		return 0, err
	}
	expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
	return moneyData.MoneyMap[expMoneyID], nil
}
