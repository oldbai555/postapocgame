package bag

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// AddItemTxUseCase 添加物品事务用例（仅更新内存状态，不触发事件）
// 用于与 MoneySys 组合的事务逻辑，确保原子性
type AddItemTxUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
}

// NewAddItemTxUseCase 创建添加物品事务用例
func NewAddItemTxUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *AddItemTxUseCase {
	return &AddItemTxUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// Execute 执行添加物品事务用例（不触发事件）
func (uc *AddItemTxUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32, bind uint32) error {
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
	var maxStack uint32 = 1
	if itemConfig, ok := itemConfigRaw.(*jsonconf.ItemConfig); ok {
		maxStack = itemConfig.MaxStack
	}

	if maxStack > 1 {
		// 可堆叠物品，尝试合并
		if indexedItems, exists := itemIndex[itemID]; exists {
			for _, existing := range indexedItems {
				if existing == nil || existing.ItemId != itemID || existing.Bind != bind {
					continue
				}
				maxAdd := maxStack - existing.Count
				if maxAdd > 0 {
					addCount := count
					if addCount > maxAdd {
						addCount = maxAdd
					}
					existing.Count += addCount
					count -= addCount
					if count == 0 {
						break
					}
				}
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
		// 更新辅助索引（注意：这里不更新外部 itemIndex，因为这是临时索引，外部调用者会重建）
		itemIndex[itemID] = append(itemIndex[itemID], newItem)
	}

	// 注意：不触发事件，这是事务型操作
	return nil
}

// getBagSize 获取背包容量（从配置读取）
func (uc *AddItemTxUseCase) getBagSize(bagType uint32) uint32 {
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
