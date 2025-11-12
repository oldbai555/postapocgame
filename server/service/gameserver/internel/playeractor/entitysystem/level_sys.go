package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// LevelSys 等级系统
type LevelSys struct {
	*BaseSystem
}

// NewLevelSys 创建等级系统
func NewLevelSys() *LevelSys {
	sys := &LevelSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysLevel)),
	}
	return sys
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSys()
	})

	// 注册玩家级别的事件处理器（这些会被克隆到每个玩家）
	gevent.SubscribePlayerEventH(gevent.OnPlayerLevelUp, func(ctx context.Context, ev *event.Event) {
	})

	gevent.SubscribePlayerEventH(gevent.OnPlayerExpChange, func(ctx context.Context, ev *event.Event) {
	})
}
