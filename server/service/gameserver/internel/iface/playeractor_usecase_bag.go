package iface

import (
	"context"
	"postapocgame/server/internal/protocol"
)

// BagUseCase 背包系统用例接口（Use Case 层定义，用于 EquipSys 依赖）
type BagUseCase interface {
	// GetItem 获取物品
	GetItem(ctx context.Context, roleID uint64, itemID uint32) (*protocol.ItemSt, error)

	// RemoveItem 移除物品
	RemoveItem(ctx context.Context, roleID uint64, itemID uint32, count uint32) error

	// AddItem 添加物品
	AddItem(ctx context.Context, roleID uint64, itemID uint32, count uint32, bind uint32) error

	// HasItem 检查是否拥有足够数量的指定物品
	HasItem(ctx context.Context, roleID uint64, itemID uint32, count uint32) (bool, error)
}
