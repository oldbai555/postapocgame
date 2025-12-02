package publicactor

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 在服务器启动事件中注册 PublicActor 内部消息处理器
func init() {
	gevent2.Subscribe(gevent2.OnSrvStart, func(ctx context.Context, e *event.Event) {
		facade := gshare.GetPublicActorFacade()
		if facade == nil {
			return
		}

		// 按业务模块拆分注册，避免在一个函数里堆所有 handler
		RegisterOnlineHandlers(facade)
		RegisterChatHandlers(facade)
		RegisterFriendHandlers(facade)
		RegisterRankHandlers(facade)
		RegisterGuildHandlers(facade)
		RegisterAuctionHandlers(facade)
		RegisterOfflineDataHandlers(facade)
	})
}
