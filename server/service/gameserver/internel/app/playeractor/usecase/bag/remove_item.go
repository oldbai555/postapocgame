package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
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

	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	remaining := count
	itemsToRemove := make([]int, 0) // 记录需要删除的索引

	// 查找需要删除的物品
	for i, item := range bagData.Items {
		if item == nil || item.ItemId != itemID {
			continue
		}
		if item.Count > remaining {
			item.Count -= remaining
			remaining = 0
			break
		}
		remaining -= item.Count
		itemsToRemove = append(itemsToRemove, i)
		if remaining == 0 {
			break
		}
	}

	if remaining > 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	// 从后往前删除，避免索引变化
	for i := len(itemsToRemove) - 1; i >= 0; i-- {
		idx := itemsToRemove[i]
		// 从切片中删除
		bagData.Items = append(bagData.Items[:idx], bagData.Items[idx+1:]...)
	}

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnItemRemove, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}
