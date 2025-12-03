package dungeonactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entity"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/fuben"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 在服务器启动事件中注册 DungeonActor 内部消息处理器
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, e *event.Event) {
		facade := gshare.GetDungeonActorFacade()
		if facade == nil {
			return
		}

		// 按业务模块拆分注册，避免在一个函数里堆所有 handler
		RegisterMoveHandlers(facade)
		RegisterFightHandlers(facade)
		RegisterFuBenHandlers(facade)
	})
}

// RegisterMoveHandlers 注册移动相关消息处理器
func RegisterMoveHandlers(facade gshare.IDungeonActorFacade) {
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdStartMove), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleStartMove(msg); err != nil {
			log.Errorf("[dungeonactor] handleStartMove failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdUpdateMove), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleUpdateMove(msg); err != nil {
			log.Errorf("[dungeonactor] handleUpdateMove failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdEndMove), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleEndMove(msg); err != nil {
			log.Errorf("[dungeonactor] handleEndMove failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdChangeScene), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleChangeScene(msg); err != nil {
			log.Errorf("[dungeonactor] handleChangeScene failed: %v", err)
		}
	})
}

// RegisterFightHandlers 注册战斗相关消息处理器
func RegisterFightHandlers(facade gshare.IDungeonActorFacade) {
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdUseSkill), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleUseSkill(msg); err != nil {
			log.Errorf("[dungeonactor] handleUseSkill failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdPickupItem), func(msg actor.IActorMessage) {
		if err := entitysystem.HandlePickupItem(msg); err != nil {
			log.Errorf("[dungeonactor] handlePickupItem failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdRevive), func(msg actor.IActorMessage) {
		if err := entity.HandleRevive(msg); err != nil {
			log.Errorf("[dungeonactor] handleRevive failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdGetNearestMonster), func(msg actor.IActorMessage) {
		if err := entitysystem.HandleGetNearestMonster(msg); err != nil {
			log.Errorf("[dungeonactor] handleGetNearestMonster failed: %v", err)
		}
	})
}

func RegisterFuBenHandlers(facade gshare.IDungeonActorFacade) {
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdEnterDungeon), func(msg actor.IActorMessage) {
		if err := fuben.HandleG2DEnterDungeon(msg); err != nil {
			log.Errorf("[dungeonactor] handleG2DEnterDungeon failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdSyncAttrs), func(msg actor.IActorMessage) {
		if err := fuben.HandleG2DSyncAttrs(msg); err != nil {
			log.Errorf("[dungeonactor] handleG2DSyncAttrs failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdUpdateHpMp), func(msg actor.IActorMessage) {
		if err := fuben.HandleG2DUpdateHpMp(msg); err != nil {
			log.Errorf("[dungeonactor] handleG2DUpdateHpMp failed: %v", err)
		}
	})
	facade.RegisterHandler(uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdUpdateSkill), func(msg actor.IActorMessage) {
		if err := fuben.HandleG2DUpdateSkill(msg); err != nil {
			log.Errorf("[dungeonactor] handleG2DUpdateSkill failed: %v", err)
		}
	})
}
