package bag

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// AddItemUseCase 添加物品用例
type AddItemUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
	configManager  interfaces.ConfigManager
}

// NewAddItemUseCase 创建添加物品用例
func NewAddItemUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
	configManager interfaces.ConfigManager,
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
	itemConfigRaw, ok := uc.configManager.GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 确保 BagData 已初始化
	if binaryData.BagData == nil {
		binaryData.BagData = &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}
	}
	if binaryData.BagData.Items == nil {
		binaryData.BagData.Items = make([]*protocol.ItemSt, 0)
	}

	// 构建辅助索引（用于快速查找）
	itemIndex := make(map[uint32][]*protocol.ItemSt)
	for _, item := range binaryData.BagData.Items {
		if item != nil {
			itemIndex[item.ItemId] = append(itemIndex[item.ItemId], item)
		}
	}

	// 检查是否可以堆叠
	// 类型断言：itemConfigRaw 应该是 *jsonconf.ItemConfig
	var maxStack uint32 = 1
	if itemConfig, ok := itemConfigRaw.(*jsonconf.ItemConfig); ok {
		maxStack = itemConfig.MaxStack
	}

	if maxStack > 1 {
		// 可堆叠物品，尝试合并
		existing := uc.findItemByKey(binaryData.BagData.Items, itemIndex, itemID, bind)
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
		if len(binaryData.BagData.Items) >= int(bagSize) {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag is full")
		}

		// 创建新物品
		newItem := &protocol.ItemSt{
			ItemId: itemID,
			Count:  count,
			Bind:   bind,
		}
		binaryData.BagData.Items = append(binaryData.BagData.Items, newItem)
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
	bagConfigRaw, ok := uc.configManager.GetBagConfig(bagType)
	if !ok || bagConfigRaw == nil {
		return 100 // 默认容量
	}
	// 类型断言：bagConfigRaw 应该是 *jsonconf.BagConfig
	if bagConfig, ok := bagConfigRaw.(*jsonconf.BagConfig); ok {
		return bagConfig.Size
	}
	return 100
}
