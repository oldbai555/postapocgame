package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

func init() {
	gevent2.Subscribe(gevent2.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		guildController := NewGuildController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SCreateGuild), guildController.HandleCreateGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SJoinGuild), guildController.HandleJoinGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLeaveGuild), guildController.HandleLeaveGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryGuildInfo), guildController.HandleQueryGuildInfo)
	})
}
