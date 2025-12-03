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
		fubenController := NewFubenController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEnterDungeon), fubenController.HandleEnterDungeon)

		// 注册 PlayerActor 消息处理器（DungeonActor → PlayerActor）
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSettleDungeon), func(message actor.IActorMessage) {
			msgCtx := message.GetContext()
			sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
			if err := fubenController.HandleSettleDungeon(msgCtx, sessionID, message.GetData()); err != nil {
				log.Errorf("[fuben-controller] HandleSettleDungeon failed: %v", err)
			}
		})
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdEnterDungeonSuccess), func(message actor.IActorMessage) {
			msgCtx := message.GetContext()
			sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
			if err := fubenController.HandleEnterDungeonSuccess(msgCtx, sessionID, message.GetData()); err != nil {
				log.Errorf("[fuben-controller] HandleEnterDungeonSuccess failed: %v", err)
			}
		})
	})
}
