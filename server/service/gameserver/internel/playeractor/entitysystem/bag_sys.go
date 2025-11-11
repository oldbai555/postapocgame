package entitysystem

import (
	"context"
	"fmt"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// BagSys 背包系统
type BagSys struct {
	*BaseSystem
	capacity uint32
	items    map[uint32]*protocol.Item // itemId -> Item
}

// NewBagSys 创建背包系统
func NewBagSys(role iface.IPlayerRole) *BagSys {
	sys := &BagSys{
		BaseSystem: NewBaseSystem(custom_id.SysBag, role),
		capacity:   50, // 初始50格
		items:      make(map[uint32]*protocol.Item),
	}
	return sys
}

// OnRoleLogin 角色登录时下发背包数据
func (s *BagSys) OnRoleLogin() {
	return
}

// SendData 下发背包数据
func (s *BagSys) SendData() error {
	items := make([]protocol.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, *item)
	}

	data := &protocol.BagData{
		Capacity: s.capacity,
		Items:    items,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(protocol.S2C_BagData, jsonData)
}

// AddItem 添加道具
func (s *BagSys) AddItem(item protocol.Item) error {
	// 检查容量
	if uint32(len(s.items)) >= s.capacity {
		return ErrBagFull
	}

	if existing, ok := s.items[item.ItemId]; ok {
		existing.Count += item.Count
	} else {
		s.items[item.ItemId] = &item
	}

	// 发布道具添加事件
	s.role.Publish(gevent.OnItemAdd, item.ItemId, item.Count)

	return s.SendData()
}

// ConsumeItem 消耗道具
func (s *BagSys) ConsumeItem(itemID uint32, count uint32) error {
	item, ok := s.items[itemID]
	if !ok {
		return ErrItemNotFound
	}

	if item.Count < count {
		return ErrItemNotEnough
	}

	item.Count -= count
	if item.Count == 0 {
		delete(s.items, itemID)
	}

	// 发布道具移除事件
	s.role.Publish(gevent.OnItemRemove, itemID, count)

	return s.SendData()
}

// HasEnough 检查是否足够
func (s *BagSys) HasEnough(itemID uint32, count uint32) bool {
	if item, ok := s.items[itemID]; ok {
		return item.Count >= count
	}
	return false
}

// ExpandCapacity 扩展容量
func (s *BagSys) ExpandCapacity(add uint32) {
	s.capacity += add

	// 发布背包扩展事件
	s.role.Publish(gevent.OnBagExpand, s.capacity, add)

	s.SendData()
}

// GetCapacity 获取容量
func (s *BagSys) GetCapacity() uint32 {
	return s.capacity
}

// GetItemCount 获取道具数量
func (s *BagSys) GetItemCount(itemID uint32) uint32 {
	if item, ok := s.items[itemID]; ok {
		return item.Count
	}
	return 0
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(custom_id.SysBag, func(role iface.IPlayerRole) iface.ISystem {
		return NewBagSys(role)
	})

	// 注册玩家级别的事件处理器
	gevent.SubscribePlayerEvent(gevent.OnItemAdd, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			itemID, _ := ev.Data[0].(uint32)
			count, _ := ev.Data[1].(uint32)
			log.Infof("[BagSys Event] Item added: itemID=%d, count=%d (source: %s)", itemID, count, ev.Source)

			// 这里可以处理道具添加后的逻辑，比如：
			// 1. 检查是否完成任务目标
			// 2. 触发成就系统
			// 3. 发送道具获得通知
		}
	})

	gevent.SubscribePlayerEvent(gevent.OnItemRemove, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			itemID, _ := ev.Data[0].(uint32)
			count, _ := ev.Data[1].(uint32)
			log.Infof("[BagSys Event] Item removed: itemID=%d, count=%d (source: %s)", itemID, count, ev.Source)
		}
	})

	gevent.SubscribePlayerEventL(gevent.OnBagExpand, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			newCapacity, _ := ev.Data[0].(uint32)
			added, _ := ev.Data[1].(uint32)
			log.Infof("[BagSys Event] Bag expanded: newCapacity=%d, added=%d (source: %s)", newCapacity, added, ev.Source)
		}
	})
}

var (
	ErrBagFull       = fmt.Errorf("bag is full")
	ErrItemNotFound  = fmt.Errorf("item not found")
	ErrItemNotEnough = fmt.Errorf("item not enough")
)
