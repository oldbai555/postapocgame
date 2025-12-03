package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// RemoveItemTxUseCase 移除物品事务用例（仅更新内存状态，不触发事件）
// 用于与 MoneySys 组合的事务逻辑，确保原子性
type RemoveItemTxUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewRemoveItemTxUseCase 创建移除物品事务用例
func NewRemoveItemTxUseCase(
	playerRepo repository.PlayerRepository,
) *RemoveItemTxUseCase {
	return &RemoveItemTxUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行移除物品事务用例（不触发事件）
func (uc *RemoveItemTxUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) error {
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

	// 注意：不触发事件，这是事务型操作
	return nil
}
