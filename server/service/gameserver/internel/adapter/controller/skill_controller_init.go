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
		skillController := NewSkillController()
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLearnSkill), skillController.HandleLearnSkill)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUpgradeSkill), skillController.HandleUpgradeSkill)
		// 技能释放（战斗服内技能使用）转发给 DungeonActor
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUseSkill), skillController.HandleUseSkill)
	})
}
