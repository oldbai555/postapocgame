package controller

import (
	"context"

	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 注册拾取掉落物相关 C2S 协议到 DungeonItemController
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		itemController := NewDungeonItemController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SPickupItem), itemController.HandlePickupItem)
	})
}
