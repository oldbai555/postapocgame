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
		auctionController := NewAuctionController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionPutOn), auctionController.HandlePutOn)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionBuy), auctionController.HandleBuy)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionQuery), auctionController.HandleQuery)
	})
}
