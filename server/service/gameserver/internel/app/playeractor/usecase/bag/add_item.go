package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// AddItemUseCase 添加物品用例
type AddItemUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces2.EventPublisher
	configManager  interfaces2.ConfigManager
}

// NewAddItemUseCase 创建添加物品用例
func NewAddItemUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces2.EventPublisher,
	configManager interfaces2.ConfigManager,
) *AddItemUseCase {
	return &AddItemUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
		configManager:  configManager,
	}
}

// Execute 执行添加物品用例
func (uc *AddItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32, bind uint32) error {
	if count == 0 {
		return nil
	}

	// 检查物品配置
	itemConfig := uc.configManager.GetItemConfig(itemID)
	if itemConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 构建辅助索引（用于快速查找）
	itemIndex := make(map[uint32][]*protocol.ItemSt)
	for _, item := range bagData.Items {
		if item != nil {
			itemIndex[item.ItemId] = append(itemIndex[item.ItemId], item)
		}
	}

	// 检查是否可以堆叠
	var maxStack uint32 = 1
	if itemConfig != nil {
		maxStack = itemConfig.MaxStack
	}

	if maxStack > 1 {
		// 可堆叠物品，尝试合并
		existing := uc.findItemByKey(bagData.Items, itemIndex, itemID, bind)
		if existing != nil {
			// 检查堆叠上限
			maxAdd := maxStack - existing.Count
			if maxAdd > 0 {
				addCount := count
				if addCount > maxAdd {
					addCount = maxAdd
				}
				existing.Count += addCount
				count -= addCount
			}
		}
	}

	// 如果还有剩余，创建新物品
	if count > 0 {
		// 检查背包容量（从配置读取）
		bagSize := uc.getBagSize(1) // 默认背包类型为1
		if len(bagData.Items) >= int(bagSize) {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag is full")
		}

		// 创建新物品
		newItem := &protocol.ItemSt{
			ItemId: itemID,
			Count:  count,
			Bind:   bind,
		}
		bagData.Items = append(bagData.Items, newItem)
		// 更新辅助索引
		itemIndex[itemID] = append(itemIndex[itemID], newItem)
	}

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnItemAdd, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}

// findItemByKey 根据itemID和bind查找物品（用于堆叠查找）
func (uc *AddItemUseCase) findItemByKey(items []*protocol.ItemSt, itemIndex map[uint32][]*protocol.ItemSt, itemID uint32, bind uint32) *protocol.ItemSt {
	if items == nil {
		return nil
	}
	// 使用辅助索引快速定位
	if indexedItems, exists := itemIndex[itemID]; exists {
		for _, item := range indexedItems {
			if item != nil && item.ItemId == itemID && item.Bind == bind {
				return item
			}
		}
	}
	return nil
}

// getBagSize 获取背包容量（从配置读取）
func (uc *AddItemUseCase) getBagSize(bagType uint32) uint32 {
	bagConfig := uc.configManager.GetBagConfig(bagType)
	if bagConfig == nil {
		return 100 // 默认容量
	}
	return bagConfig.Size
}
