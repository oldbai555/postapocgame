package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// BagSys 背包系统
type BagSys struct {
	*BaseSystem
}

// NewBagSys 创建背包系统
func NewBagSys() *BagSys {
	sys := &BagSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysBag)),
	}
	return sys
}

// OnRoleLogin 角色登录时下发背包数据
func (s *BagSys) OnRoleLogin() {
	return
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
