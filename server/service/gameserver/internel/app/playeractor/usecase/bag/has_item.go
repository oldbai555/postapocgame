package bag

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
)

// HasItemUseCase 检查物品用例
type HasItemUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewHasItemUseCase 创建检查物品用例
func NewHasItemUseCase(playerRepo repository.PlayerRepository) *HasItemUseCase {
	return &HasItemUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行检查物品用例
func (uc *HasItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) (bool, error) {
	if count == 0 {
		return true, nil
	}

	acc, err := newAccessor(ctx, uc.playerRepo)
	if err != nil {
		return false, err
	}
	return acc.totalCount(itemID) >= count, nil
}
