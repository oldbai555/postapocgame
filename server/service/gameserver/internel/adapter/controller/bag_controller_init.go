package controller

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	gshare2 "postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/di"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

func init() {
	gevent2.Subscribe(gevent2.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		bagController := NewBagController()

		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SOpenBag), bagController.HandleOpenBag)
		di.GetContainer().DungeonServerGateway().RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GAddItem), bagController.HandleAddItem)

		gshare2.RegisterHandler(uint16(protocol.D2GRpcProtocol_D2GAddItem), func(message actor.IActorMessage) {
			msgCtx := message.GetContext()
			sessionID, _ := msgCtx.Value(gshare2.ContextKeySession).(string)
			_ = bagController.HandleAddItem(msgCtx, sessionID, message.GetData())
		})
	})
}
