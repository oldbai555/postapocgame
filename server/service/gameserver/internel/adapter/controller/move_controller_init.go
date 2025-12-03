package controller

import (
	"context"

	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 注册移动相关 C2S 协议到 MoveController
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		moveController := NewMoveController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SStartMove), moveController.HandleStartMove)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUpdateMove), moveController.HandleUpdateMove)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEndMove), moveController.HandleEndMove)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SChangeScene), moveController.HandleChangeScene)
	})
}
