package entitysystem

import (
	"fmt"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/tool"
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
	return s.role.SendMessage(1, 7, jsonData)
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
}

var (
	ErrBagFull       = fmt.Errorf("bag is full")
	ErrItemNotFound  = fmt.Errorf("item not found")
	ErrItemNotEnough = fmt.Errorf("item not enough")
)
