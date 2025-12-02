package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
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

	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	if binaryData.BagData == nil || binaryData.BagData.Items == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	remaining := count
	itemsToRemove := make([]int, 0) // 记录需要删除的索引

	// 查找需要删除的物品
	for i, item := range binaryData.BagData.Items {
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
		binaryData.BagData.Items = append(binaryData.BagData.Items[:idx], binaryData.BagData.Items[idx+1:]...)
	}

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnItemRemove, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}
