package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// BagSys 背包系统
type BagSys struct {
	*BaseSystem
	bagData *protocol.SiBagData
}

// NewBagSys 创建背包系统
func NewBagSys() *BagSys {
	return &BagSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysBag)),
	}
}

func GetBagSys(ctx context.Context) *BagSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysBag))
	if system == nil {
		log.Errorf("not found system [%v] error:%v", protocol.SystemId_SysBag, err)
		return nil
	}
	sys := system.(*BagSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error:%v", protocol.SystemId_SysBag, err)
		return nil
	}
	return sys
}

// OnInit 初始化时从数据库加载背包数据
func (bs *BagSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果bag_data不存在，则初始化
	if binaryData.BagData == nil {
		binaryData.BagData = &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}
	}
	bs.bagData = binaryData.BagData

	log.Infof("BagSys initialized: ItemCount=%d", len(bs.bagData.Items))
}

// findItemByKey 根据itemID和bind查找物品（用于堆叠查找）
func (bs *BagSys) findItemByKey(itemID uint32, bind uint32) *protocol.ItemSt {
	if bs.bagData == nil || bs.bagData.Items == nil {
		return nil
	}
	for _, item := range bs.bagData.Items {
		if item != nil && item.ItemId == itemID && item.Bind == bind {
			return item
		}
	}
	return nil
}

// AddItem 添加物品
func (bs *BagSys) AddItem(ctx context.Context, itemID uint32, count uint32, bind uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 检查物品配置
	itemConfig, ok := jsonconf.GetConfigManager().GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	// 确保bagData已初始化
	if bs.bagData == nil {
		binaryData := playerRole.GetBinaryData()
		if binaryData.BagData == nil {
			binaryData.BagData = &protocol.SiBagData{
				Items: make([]*protocol.ItemSt, 0),
			}
		}
		bs.bagData = binaryData.BagData
	}
	if bs.bagData.Items == nil {
		bs.bagData.Items = make([]*protocol.ItemSt, 0)
	}

	// 检查是否可以堆叠
	if itemConfig.MaxStack > 1 {
		// 可堆叠物品，尝试合并
		existing := bs.findItemByKey(itemID, bind)
		if existing != nil {
			// 检查堆叠上限
			maxAdd := itemConfig.MaxStack - existing.Count
			if maxAdd > 0 {
				addCount := count
				if addCount > maxAdd {
					addCount = maxAdd
				}
				existing.Count += addCount
				count -= addCount

				// 不再需要数据库操作，数据已存储在BinaryData中
			}
		}
	}

	// 如果还有剩余，创建新物品
	if count > 0 {
		// 检查背包容量（简单实现，可以扩展）
		if len(bs.bagData.Items) >= 100 {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag is full")
		}

		// 添加到内存（不再需要数据库操作）
		newItem := &protocol.ItemSt{
			ItemId: itemID,
			Count:  count,
			Bind:   bind,
		}
		bs.bagData.Items = append(bs.bagData.Items, newItem)
	}

	// 发布事件
	playerRole.Publish(gevent.OnItemAdd, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}

// HasItem 检查是否拥有足够数量的指定物品
func (bs *BagSys) HasItem(itemID uint32, count uint32) bool {
	if count == 0 {
		return true
	}
	if bs.bagData == nil || bs.bagData.Items == nil {
		return false
	}
	var total uint32
	for _, item := range bs.bagData.Items {
		if item != nil && item.ItemId == itemID {
			total += item.Count
			if total >= count {
				return true
			}
		}
	}
	return false
}

// RemoveItem 移除物品
func (bs *BagSys) RemoveItem(ctx context.Context, itemID uint32, count uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	if bs.bagData == nil || bs.bagData.Items == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	remaining := count
	itemsToRemove := make([]int, 0) // 记录需要删除的索引
	for i, v := range bs.bagData.Items {
		if v == nil || v.ItemId != itemID {
			continue
		}
		if v.Count > remaining {
			v.Count -= remaining
			remaining = 0
			break
		}
		remaining -= v.Count
		itemsToRemove = append(itemsToRemove, i)
		if remaining == 0 {
			break
		}
	}
	if remaining > 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	// 从后往前删除，避免索引变化（不再需要数据库操作）
	for i := len(itemsToRemove) - 1; i >= 0; i-- {
		idx := itemsToRemove[i]
		// 从切片中删除
		bs.bagData.Items = append(bs.bagData.Items[:idx], bs.bagData.Items[idx+1:]...)
	}

	// 发布事件
	playerRole.Publish(gevent.OnItemRemove, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}

// GetItem 获取物品
func (bs *BagSys) GetItem(itemID uint32) *protocol.ItemSt {
	if bs.bagData == nil || bs.bagData.Items == nil {
		return nil
	}
	for _, item := range bs.bagData.Items {
		if item != nil && item.ItemId == itemID {
			return item
		}
	}
	return nil
}

// GetAllItems 获取所有物品
func (bs *BagSys) GetAllItems() []*protocol.ItemSt {
	if bs.bagData == nil || bs.bagData.Items == nil {
		return make([]*protocol.ItemSt, 0)
	}
	return bs.bagData.Items
}

// GetBagData 获取背包数据（用于协议）
func (bs *BagSys) GetBagData() *protocol.SiBagData {
	return bs.bagData
}

// RemoveItemTx 移除物品（支持事务，仅更新内存状态，不写数据库）
func (bs *BagSys) RemoveItemTx(itemID uint32, count uint32) error {
	if bs.bagData == nil || bs.bagData.Items == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	remaining := count
	itemsToRemove := make([]int, 0) // 记录需要删除的索引
	for i, v := range bs.bagData.Items {
		if v == nil || v.ItemId != itemID {
			continue
		}
		if v.Count > remaining {
			v.Count -= remaining
			remaining = 0
			break
		}
		remaining -= v.Count
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
		bs.bagData.Items = append(bs.bagData.Items[:idx], bs.bagData.Items[idx+1:]...)
	}
	return nil
}

// AddItemTx 添加物品（支持事务，仅更新内存状态，不写数据库）
func (bs *BagSys) AddItemTx(itemID uint32, count uint32, bind uint32) error {
	if bs.bagData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag data not initialized")
	}
	if bs.bagData.Items == nil {
		bs.bagData.Items = make([]*protocol.ItemSt, 0)
	}

	// 检查物品配置
	itemConfig, ok := jsonconf.GetConfigManager().GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	// 检查是否可以堆叠
	if itemConfig.MaxStack > 1 {
		// 可堆叠物品，尝试合并
		existing := bs.findItemByKey(itemID, bind)
		if existing != nil {
			// 检查堆叠上限
			maxAdd := itemConfig.MaxStack - existing.Count
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
		// 检查背包容量（默认100格）
		if len(bs.bagData.Items) >= 100 {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag is full")
		}

		newItem := &protocol.ItemSt{
			ItemId: itemID,
			Count:  count,
			Bind:   bind,
		}
		bs.bagData.Items = append(bs.bagData.Items, newItem)
	}

	return nil
}

// GetItemsSnapshot 获取物品快照（用于事务回滚）
func (bs *BagSys) GetItemsSnapshot() map[uint32]*protocol.ItemSt {
	snapshot := make(map[uint32]*protocol.ItemSt)
	if bs.bagData != nil && bs.bagData.Items != nil {
		for _, item := range bs.bagData.Items {
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
	return snapshot
}

// RestoreItemsSnapshot 恢复物品快照（用于事务回滚）
func (bs *BagSys) RestoreItemsSnapshot(snapshot map[uint32]*protocol.ItemSt) {
	if bs.bagData == nil {
		bs.bagData = &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}
	}
	// 重建Items
	bs.bagData.Items = make([]*protocol.ItemSt, 0, len(snapshot))
	for _, v := range snapshot {
		if v != nil {
			item := &protocol.ItemSt{
				ItemId: v.ItemId,
				Count:  v.Count,
				Bind:   v.Bind,
			}
			bs.bagData.Items = append(bs.bagData.Items, item)
		}
	}
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysBag), func() iface.ISystem {
		return NewBagSys()
	})
	gevent.SubscribePlayerEvent(gevent.OnItemAdd, func(ctx context.Context, ev *event.Event) {})
	gevent.SubscribePlayerEvent(gevent.OnItemRemove, func(ctx context.Context, ev *event.Event) {})
	gevent.SubscribePlayerEventL(gevent.OnBagExpand, func(ctx context.Context, ev *event.Event) {})
}
