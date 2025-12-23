package dungeonactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/dungeonactor/entitysystem"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
)

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, e *event.Event) {
		facade := gshare.GetDungeonActorFacade()
		if facade == nil {
			return
		}

		RegisterEnterGameHandler(facade)
		RegisterMoveHandlers(facade)
		RegisterFightHandlers(facade)
	})
}

func RegisterEnterGameHandler(facade gshare.IDungeonActorFacade) {
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DAMEnterGame), func(msg actor.IActorMessage) {
		if err := handleEnterGame(msg); err != nil {
			log.Errorf("[dungeon-actor] handleEnterGame failed: %v", err)
		}
	})
}

func RegisterMoveHandlers(facade gshare.IDungeonActorFacade) {
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DAMStartMove), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleStartMove(msg); err != nil {
			log.Errorf("[dungeon-actor] handleStartMove failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DAMUpdateMove), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleUpdateMove(msg); err != nil {
			log.Errorf("[dungeon-actor] handleUpdateMove failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DAMEndMove), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleEndMove(msg); err != nil {
			log.Errorf("[dungeon-actor] handleEndMove failed: %v", err)
		}
	})
}

func RegisterFightHandlers(facade gshare.IDungeonActorFacade) {
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DAMUseSkill), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleUseSkill(msg); err != nil {
			log.Errorf("[dungeon-actor] handleUseSkill failed: %v", err)
		}
	})
}
