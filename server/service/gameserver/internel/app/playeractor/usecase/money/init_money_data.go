package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
)

// InitMoneyDataUseCase 初始化货币数据用例
// 负责货币数据的初始化（MoneyData 结构、默认金币注入）
type InitMoneyDataUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewInitMoneyDataUseCase 创建初始化货币数据用例
func NewInitMoneyDataUseCase(
	playerRepo repository.PlayerRepository,
) *InitMoneyDataUseCase {
	return &InitMoneyDataUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行初始化货币数据用例
func (uc *InitMoneyDataUseCase) Execute(ctx context.Context, roleID uint64) error {
	moneyData, err := uc.playerRepo.GetMoneyData(ctx)
	if err != nil {
		return err
	}

	// 如果MoneyMap为空，初始化默认金币
	if len(moneyData.MoneyMap) == 0 {
		defaultGoldMoneyID := uint32(protocol.MoneyType_MoneyTypeGoldCoin)
		defaultGoldAmount := int64(100000)
		moneyData.MoneyMap[defaultGoldMoneyID] = defaultGoldAmount
	}

	return nil
}
