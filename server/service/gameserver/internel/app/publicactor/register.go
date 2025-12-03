package publicactor

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 在服务器启动事件中注册 PublicActor 内部消息处理器
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, e *event.Event) {
		facade := gshare.GetPublicActorFacade()
		if facade == nil {
			return
		}

		// 按业务模块拆分注册，避免在一个函数里堆所有 handler
		RegisterOnlineHandlers(facade)
		RegisterChatHandlers(facade)
		RegisterRankHandlers(facade)
		RegisterOfflineDataHandlers(facade)
	})
}
