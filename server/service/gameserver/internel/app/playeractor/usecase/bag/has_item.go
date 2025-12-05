package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
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

	// 获取 BinaryData（共享引用）
	bagData, err := uc.playerRepo.GetBagData(ctx)
	if err != nil {
		return false, err
	}
	// 构建辅助索引（用于快速查找）
	itemIndex := make(map[uint32][]*protocol.ItemSt)
	for _, item := range bagData.Items {
		if item != nil {
			itemIndex[item.ItemId] = append(itemIndex[item.ItemId], item)
		}
	}
	// 使用辅助索引快速定位
	var total uint32
	if items, exists := itemIndex[itemID]; exists {
		for _, item := range items {
			if item != nil {
				total += item.Count
				if total >= count {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
