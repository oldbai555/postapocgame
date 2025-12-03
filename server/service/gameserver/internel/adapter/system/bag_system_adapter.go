package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/bag"
)

// BagSystemAdapter 背包系统适配器
//
// 生命周期职责：
// - OnInit: 初始化 BagData 结构（如果不存在），重建辅助索引 itemIndex
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（添加物品、移除物品、堆叠规则、容量检查）均在 UseCase 层实现
// 状态管理：维护 itemIndex 辅助索引，用于快速查找物品（数据源仍为 BagData.Items）
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type BagSystemAdapter struct {
	*BaseSystemAdapter
	addItemUseCase      *bag.AddItemUseCase
	removeItemUseCase   *bag.RemoveItemUseCase
	addItemTxUseCase    *bag.AddItemTxUseCase
	removeItemTxUseCase *bag.RemoveItemTxUseCase
	hasItemUseCase      *bag.HasItemUseCase // TODO(adapter-phaseA-A2): 纯校验型逻辑后续可统一通过 BagUseCase 接口暴露，SystemAdapter 只做路由
	// 辅助索引：itemID -> []*ItemSt（用于快速查找，但不作为数据源）
	// 注意：这个索引只用于查找优化，数据源仍然是bagData.Items
	itemIndex map[uint32][]*protocol.ItemSt
}

// NewBagSystemAdapter 创建背包系统适配器
func NewBagSystemAdapter() *BagSystemAdapter {
	container := di.GetContainer()
	return &BagSystemAdapter{
		BaseSystemAdapter:   NewBaseSystemAdapter(uint32(protocol.SystemId_SysBag)),
		addItemUseCase:      bag.NewAddItemUseCase(container.PlayerGateway(), container.EventPublisher(), container.ConfigGateway()),
		removeItemUseCase:   bag.NewRemoveItemUseCase(container.PlayerGateway(), container.EventPublisher()),
		addItemTxUseCase:    bag.NewAddItemTxUseCase(container.PlayerGateway(), container.ConfigGateway()),
		removeItemTxUseCase: bag.NewRemoveItemTxUseCase(container.PlayerGateway()),
		hasItemUseCase:      bag.NewHasItemUseCase(container.PlayerGateway()),
		itemIndex:           make(map[uint32][]*protocol.ItemSt),
	}
}

// OnInit 系统初始化
func (a *BagSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("bag sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("bag sys OnInit get binary data err:%v", err)
		return
	}

	// 如果bag_data不存在，则初始化
	if binaryData.BagData == nil {
		binaryData.BagData = &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}
	}

	// 初始化辅助索引
	a.rebuildIndex(binaryData.BagData)

	log.Infof("BagSys initialized: ItemCount=%d", len(binaryData.BagData.Items))
}

// rebuildIndex 重建辅助索引（在数据变更后调用）
func (a *BagSystemAdapter) rebuildIndex(bagData *protocol.SiBagData) {
	a.itemIndex = make(map[uint32][]*protocol.ItemSt)
	if bagData == nil || bagData.Items == nil {
		return
	}
	for _, item := range bagData.Items {
		if item != nil {
			a.itemIndex[item.ItemId] = append(a.itemIndex[item.ItemId], item)
		}
	}
}

// AddItem 添加物品（对外接口，供其他系统调用）
func (a *BagSystemAdapter) AddItem(ctx context.Context, itemID uint32, count uint32, bind uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	err = a.addItemUseCase.Execute(ctx, roleID, itemID, count, bind)
	if err != nil {
		return err
	}
	// 重建辅助索引
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err == nil && binaryData != nil && binaryData.BagData != nil {
		a.rebuildIndex(binaryData.BagData)
	}
	return nil
}

// RemoveItem 移除物品（对外接口，供其他系统调用）
func (a *BagSystemAdapter) RemoveItem(ctx context.Context, itemID uint32, count uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	err = a.removeItemUseCase.Execute(ctx, roleID, itemID, count)
	if err != nil {
		return err
	}
	// 重建辅助索引
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err == nil && binaryData != nil && binaryData.BagData != nil {
		a.rebuildIndex(binaryData.BagData)
	}
	return nil
}

// HasItem 检查是否拥有足够数量的指定物品
func (a *BagSystemAdapter) HasItem(ctx context.Context, itemID uint32, count uint32) (bool, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return false, err
	}
	return a.hasItemUseCase.Execute(ctx, roleID, itemID, count)
}

// GetItem 获取物品（使用辅助索引优化查找）
func (a *BagSystemAdapter) GetItem(ctx context.Context, itemID uint32) (*protocol.ItemSt, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.BagData == nil || binaryData.BagData.Items == nil {
		return nil, nil
	}
	// 使用辅助索引快速定位
	if items, exists := a.itemIndex[itemID]; exists && len(items) > 0 {
		return items[0], nil // 返回第一个匹配的物品
	}
	return nil, nil
}

// GetAllItems 获取所有物品
func (a *BagSystemAdapter) GetAllItems(ctx context.Context) ([]*protocol.ItemSt, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.BagData == nil || binaryData.BagData.Items == nil {
		return make([]*protocol.ItemSt, 0), nil
	}
	return binaryData.BagData.Items, nil
}

// GetItemsSnapshot 获取物品快照（用于事务回滚）
// 说明：该方法主要用于兼容旧的经济系统逻辑（ApplyConsume/GrantRewards），不会触发事件，
// 仅操作 BinaryData 中的内存数据，符合当前“数据只存 BinaryData、不直接访问数据库”的约束。
func (a *BagSystemAdapter) GetItemsSnapshot(ctx context.Context) (map[uint32]*protocol.ItemSt, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	snapshot := make(map[uint32]*protocol.ItemSt)
	if binaryData.BagData != nil && binaryData.BagData.Items != nil {
		for _, item := range binaryData.BagData.Items {
			if item != nil {
				key := item.ItemId*1000 + item.Bind
				snapshot[key] = &protocol.ItemSt{
					ItemId: item.ItemId,
					Count:  item.Count,
					Bind:   item.Bind,
				}
			}
		}
	}
	return snapshot, nil
}

// RestoreItemsSnapshot 恢复物品快照（用于事务回滚）
func (a *BagSystemAdapter) RestoreItemsSnapshot(ctx context.Context, snapshot map[uint32]*protocol.ItemSt) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}
	if binaryData.BagData == nil {
		binaryData.BagData = &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}
	}
	// 重建 Items
	binaryData.BagData.Items = make([]*protocol.ItemSt, 0, len(snapshot))
	for _, v := range snapshot {
		if v != nil {
			item := &protocol.ItemSt{
				ItemId: v.ItemId,
				Count:  v.Count,
				Bind:   v.Bind,
			}
			binaryData.BagData.Items = append(binaryData.BagData.Items, item)
		}
	}
	// 重建辅助索引
	a.rebuildIndex(binaryData.BagData)
	return nil
}

// RemoveItemTx 移除物品（仅更新内存状态，不触发事件，用于与 MoneySys 组合的事务逻辑）
func (a *BagSystemAdapter) RemoveItemTx(ctx context.Context, itemID uint32, count uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	err = a.removeItemTxUseCase.Execute(ctx, roleID, itemID, count)
	if err != nil {
		return err
	}
	// 重建辅助索引（UseCase 只操作内存数据，索引维护由适配层负责）
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err == nil && binaryData != nil && binaryData.BagData != nil {
		a.rebuildIndex(binaryData.BagData)
	}
	return nil
}

// AddItemTx 添加物品（仅更新内存状态，不触发事件，用于与 MoneySys 组合的事务逻辑）
func (a *BagSystemAdapter) AddItemTx(ctx context.Context, itemID uint32, count uint32, bind uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	err = a.addItemTxUseCase.Execute(ctx, roleID, itemID, count, bind)
	if err != nil {
		return err
	}
	// 重建辅助索引（UseCase 只操作内存数据，索引维护由适配层负责）
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err == nil && binaryData != nil && binaryData.BagData != nil {
		a.rebuildIndex(binaryData.BagData)
	}
	return nil
}

// GetBagData 获取背包数据（用于协议）
func (a *BagSystemAdapter) GetBagData(ctx context.Context) (*protocol.SiBagData, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.BagData == nil {
		return &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}, nil
	}
	return binaryData.BagData, nil
}

// EnsureISystem 确保 BagSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*BagSystemAdapter)(nil)
