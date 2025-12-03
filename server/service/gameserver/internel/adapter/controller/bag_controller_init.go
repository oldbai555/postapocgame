package controller

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		bagController := NewBagController()

		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SOpenBag), bagController.HandleOpenBag)

		// 注册 PlayerActor 消息处理器（DungeonActor → PlayerActor）
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdAddItem), func(message actor.IActorMessage) {
			msgCtx := message.GetContext()
			sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
			if err := bagController.HandleAddItem(msgCtx, sessionID, message.GetData()); err != nil {
				log.Errorf("[bag-controller] HandleAddItem failed: %v", err)
			}
		})
	})
}
