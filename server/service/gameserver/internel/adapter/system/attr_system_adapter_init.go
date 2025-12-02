package system

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 注册系统工厂
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysAttr), func() iface.ISystem {
		return NewAttrSystemAdapter()
	})

	// 订阅服务器启动事件（AttrSys 不需要客户端协议，主要通过 RunOne 和事件触发）
	gevent2.Subscribe(gevent2.OnSrvStart, func(ctx context.Context, event *event.Event) {
		// AttrSys 不需要注册协议处理器
	})
}
