package bag

import (
	"context"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// RemoveItemUseCase 移除物品用例
type RemoveItemUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
}

// NewRemoveItemUseCase 创建移除物品用例
func NewRemoveItemUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
) *RemoveItemUseCase {
	return &RemoveItemUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute 执行移除物品用例
func (uc *RemoveItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) error {
	if count == 0 {
		return nil
	}

	acc, err := newAccessor(ctx, uc.playerRepo)
	if err != nil {
		return customerr.Wrap(err)
	}
	if err := acc.removeItem(itemID, count); err != nil {
		return err
	}

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnItemRemove, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}
