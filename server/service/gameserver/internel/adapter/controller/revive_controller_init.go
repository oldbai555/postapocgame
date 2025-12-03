package controller

import (
	"context"

	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 注册复活相关 C2S 协议到 ReviveController
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		reviveController := NewReviveController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRevive), reviveController.HandleRevive)
	})
}
