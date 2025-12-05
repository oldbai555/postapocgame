package interfaces

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
)

// AttrUseCase 属性系统用例接口（Use Case 层定义，用于 LevelSys 依赖）
// 注意：此接口在 AttrSys 重构后会被实现
type AttrUseCase interface {
	// MarkDirty 标记需要重算的系统
	MarkDirty(ctx context.Context, roleID uint64, sysID uint32) error
}

// IAttrCalculator 属性计算器接口（复用 core 层定义）
type IAttrCalculator = iface.IAttrCalculator
