package money

import (
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Deps 聚合 Money 相关依赖
type Deps struct {
	PlayerRepo     iface.PlayerRepository
	EventPublisher iface.EventPublisher
	NetworkGateway iface.NetworkGateway
}

// depsFromRuntime 从 Runtime 组装 Money 所需依赖
func depsFromRuntime(rt *runtime.Runtime) Deps {
	return Deps{
		PlayerRepo:     rt.PlayerRepo(),
		EventPublisher: rt.EventPublisher(),
		NetworkGateway: rt.NetworkGateway(),
	}
}
