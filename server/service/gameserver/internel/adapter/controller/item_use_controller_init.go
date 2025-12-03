package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		itemUseController := NewItemUseController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUseItem), itemUseController.HandleUseItem)
	})
}
