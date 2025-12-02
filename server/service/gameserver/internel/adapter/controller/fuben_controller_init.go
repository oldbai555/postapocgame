package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/di"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

func init() {
	gevent2.Subscribe(gevent2.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		fubenController := NewFubenController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEnterDungeon), fubenController.HandleEnterDungeon)

		dungeonGateway := di.GetContainer().DungeonServerGateway()
		dungeonGateway.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GSettleDungeon), fubenController.HandleSettleDungeon)
		dungeonGateway.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GEnterDungeonSuccess), fubenController.HandleEnterDungeonSuccess)
	})
}
