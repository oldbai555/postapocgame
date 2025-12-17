package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// AddItemUseCase 添加物品用例（Phase2B：小 service，内部使用 Deps 聚合依赖）。
type AddItemUseCase struct {
	deps Deps
}

// NewAddItemUseCase 创建添加物品用例。
func NewAddItemUseCase(deps Deps) *AddItemUseCase {
	return &AddItemUseCase{deps: deps}
}

// Execute 执行添加物品用例
func (uc *AddItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32, bind uint32) error {
	if count == 0 {
		return nil
	}

	// 检查物品配置
	itemConfig := uc.deps.ConfigManager.GetItemConfig(itemID)
	if itemConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	acc, err := newAccessor(ctx, uc.deps.PlayerRepo)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 检查是否可以堆叠
	var maxStack uint32 = 1
	if itemConfig != nil {
		maxStack = itemConfig.MaxStack
	}

	bagSize := uc.getBagSize(1) // 默认背包类型为1
	if err := acc.addItem(itemID, bind, count, maxStack, bagSize); err != nil {
		return err
	}

	// 发布事件
	uc.deps.EventPublisher.PublishPlayerEvent(ctx, gevent.OnItemAdd, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}

// getBagSize 获取背包容量（从配置读取）
func (uc *AddItemUseCase) getBagSize(bagType uint32) uint32 {
	bagConfig := uc.deps.ConfigManager.GetBagConfig(bagType)
	if bagConfig == nil {
		return 100 // 默认容量
	}
	return bagConfig.Size
}
