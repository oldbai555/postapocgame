package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

const (
	// 默认背包类型（主背包）
	DefaultBagType uint32 = 1
	// 默认背包容量（如果配置不存在时使用）
	DefaultBagSize uint32 = 100
)

// BagSys 背包系统
type BagSys struct {
	*BaseSystem
	bagData *protocol.SiBagData
	// 辅助索引：itemID -> []*ItemSt（用于快速查找，但不作为数据源）
	// 注意：这个索引只用于查找优化，数据源仍然是bagData.Items
	itemIndex map[uint32][]*protocol.ItemSt
}

// NewBagSys 创建背包系统
func NewBagSys() *BagSys {
	return &BagSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysBag)),
		itemIndex:  make(map[uint32][]*protocol.ItemSt),
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

	// 初始化辅助索引
	bs.rebuildIndex()

	log.Infof("BagSys initialized: ItemCount=%d", len(bs.bagData.Items))
}

// getBagSize 获取背包容量（从配置读取）
func (bs *BagSys) getBagSize(bagType uint32) uint32 {
	configMgr := jsonconf.GetConfigManager()
	bagConfig, ok := configMgr.GetBagConfig(bagType)
	if !ok || bagConfig == nil {
		// 如果配置不存在，使用默认值
		return DefaultBagSize
	}
	return bagConfig.Size
}

// rebuildIndex 重建辅助索引（在数据变更后调用）
func (bs *BagSys) rebuildIndex() {
	bs.itemIndex = make(map[uint32][]*protocol.ItemSt)
	if bs.bagData == nil || bs.bagData.Items == nil {
		return
	}
	for _, item := range bs.bagData.Items {
		if item != nil {
			bs.itemIndex[item.ItemId] = append(bs.itemIndex[item.ItemId], item)
		}
	}
}

// findItemByKey 根据itemID和bind查找物品（用于堆叠查找）
// 使用辅助索引优化查找效率
func (bs *BagSys) findItemByKey(itemID uint32, bind uint32) *protocol.ItemSt {
	if bs.bagData == nil || bs.bagData.Items == nil {
		return nil
	}
	// 使用辅助索引快速定位
	if items, exists := bs.itemIndex[itemID]; exists {
		for _, item := range items {
			if item != nil && item.ItemId == itemID && item.Bind == bind {
				return item
			}
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
		// 检查背包容量（从配置读取）
		bagSize := bs.getBagSize(DefaultBagType)
		if len(bs.bagData.Items) >= int(bagSize) {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag is full")
		}

		// 添加到内存（不再需要数据库操作）
		newItem := &protocol.ItemSt{
			ItemId: itemID,
			Count:  count,
			Bind:   bind,
		}
		bs.bagData.Items = append(bs.bagData.Items, newItem)
		// 更新辅助索引
		bs.itemIndex[itemID] = append(bs.itemIndex[itemID], newItem)
	}

	// 发布事件
	playerRole.Publish(gevent.OnItemAdd, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}

// HasItem 检查是否拥有足够数量的指定物品（使用辅助索引优化查找）
func (bs *BagSys) HasItem(itemID uint32, count uint32) bool {
	if count == 0 {
		return true
	}
	if bs.bagData == nil || bs.bagData.Items == nil {
		return false
	}
	var total uint32
	// 使用辅助索引快速定位
	if items, exists := bs.itemIndex[itemID]; exists {
		for _, item := range items {
			if item != nil {
				total += item.Count
				if total >= count {
					return true
				}
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

	// 重建辅助索引
	bs.rebuildIndex()

	// 发布事件
	playerRole.Publish(gevent.OnItemRemove, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}

// GetItem 获取物品（使用辅助索引优化查找）
func (bs *BagSys) GetItem(itemID uint32) *protocol.ItemSt {
	if bs.bagData == nil || bs.bagData.Items == nil {
		return nil
	}
	// 使用辅助索引快速定位
	if items, exists := bs.itemIndex[itemID]; exists && len(items) > 0 {
		return items[0] // 返回第一个匹配的物品
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
	// 重建辅助索引
	bs.rebuildIndex()
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
		// 检查背包容量（从配置读取）
		bagSize := bs.getBagSize(DefaultBagType)
		if len(bs.bagData.Items) >= int(bagSize) {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag is full")
		}

		newItem := &protocol.ItemSt{
			ItemId: itemID,
			Count:  count,
			Bind:   bind,
		}
		bs.bagData.Items = append(bs.bagData.Items, newItem)
		// 更新辅助索引
		bs.itemIndex[itemID] = append(bs.itemIndex[itemID], newItem)
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
	// 重建辅助索引
	bs.rebuildIndex()
}

func handleOpenBag(ctx context.Context, _ *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	bagSys := GetBagSys(ctx)
	var bagData *protocol.SiBagData
	if bagSys != nil {
		bagData = bagSys.GetBagData()
	} else {
		bagData = &protocol.SiBagData{}
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CBagData), &protocol.S2CBagDataReq{
		BagData: bagData,
	})
}

func pushBagData(ctx context.Context, sessionId string) {
	bagSys := GetBagSys(ctx)
	if bagSys == nil {
		return
	}
	if err := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CBagData), &protocol.S2CBagDataReq{
		BagData: bagSys.GetBagData(),
	}); err != nil {
		log.Errorf("push bag data failed: %v", err)
	}
}

func handleAddItemActorMessage(message actor.IActorMessage) {
	ctx := message.GetContext()
	sessionId, _ := ctx.Value(gshare.ContextKeySession).(string)
	if err := handleAddItem(ctx, sessionId, message.GetData()); err != nil {
		log.Errorf("handleAddItem failed: %v", err)
	}
}

// handleAddItem 处理添加物品请求（拾取掉落物）- 在Actor中异步处理
func handleAddItem(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GAddItemReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal add item request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("received add item request: RoleId=%d, ItemId=%d, Count=%d", req.RoleId, req.ItemId, req.Count)

	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Errorf("player role not found: RoleId=%d, SessionId=%s", req.RoleId, sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	roleCtx := playerRole.WithContext(ctx)
	itemCfg, ok := jsonconf.GetConfigManager().GetItemConfig(req.ItemId)
	if !ok {
		log.Errorf("item config not found: ItemId=%d", req.ItemId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found")
	}

	rewards := []*jsonconf.ItemAmount{
		{
			ItemType: itemCfg.Type,
			ItemId:   req.ItemId,
			Count:    int64(req.Count),
			Bind:     0,
		},
	}

	if err := playerRole.GrantRewards(roleCtx, rewards); err != nil {
		log.Errorf("grant rewards failed: RoleId=%d, ItemId=%d, Count=%d, Error=%v", req.RoleId, req.ItemId, req.Count, err)
		_ = playerRole.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "拾取失败，请稍后重试",
			ItemHdl: req.ItemHdl,
		})
		return customerr.Wrap(err)
	}

	if bagSys := GetBagSys(roleCtx); bagSys != nil {
		if err := playerRole.SendProtoMessage(uint16(protocol.S2CProtocol_S2CBagData), &protocol.S2CBagDataReq{
			BagData: bagSys.GetBagData(),
		}); err != nil {
			log.Warnf("send bag data failed: %v", err)
		}
	}

	log.Infof("item added successfully: RoleId=%d, ItemId=%d, Count=%d", req.RoleId, req.ItemId, req.Count)
	return nil
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysBag), func() iface.ISystem {
		return NewBagSys()
	})
	gevent.SubscribePlayerEvent(gevent.OnItemAdd, func(ctx context.Context, ev *event.Event) {})
	gevent.SubscribePlayerEvent(gevent.OnItemRemove, func(ctx context.Context, ev *event.Event) {})
	gevent.SubscribePlayerEventL(gevent.OnBagExpand, func(ctx context.Context, ev *event.Event) {})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SOpenBag), handleOpenBag)
		gshare.RegisterHandler(uint16(protocol.D2GRpcProtocol_D2GAddItem), handleAddItemActorMessage)
		// 注册添加物品的RPC处理器
		dungeonserverlink.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GAddItem), handleAddItem)
	})
}
