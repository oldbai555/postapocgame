package system

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysBag), func() iface.ISystem {
		return NewBagSystemAdapter()
	})

	// 订阅玩家事件（订阅到全局模板）
	gevent2.SubscribePlayerEvent(gevent2.OnItemAdd, func(ctx context.Context, ev *event.Event) {})
	gevent2.SubscribePlayerEvent(gevent2.OnItemRemove, func(ctx context.Context, ev *event.Event) {})
	gevent2.SubscribePlayerEventL(gevent2.OnBagExpand, func(ctx context.Context, ev *event.Event) {})

	// 协议注册由 controller 包负责，避免系统与控制器之间的循环依赖
}
