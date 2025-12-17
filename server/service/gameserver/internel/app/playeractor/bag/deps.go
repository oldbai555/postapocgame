package bag

import (
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Deps 聚合 Bag 相关依赖
type Deps struct {
	PlayerRepo     iface.PlayerRepository
	EventPublisher iface.EventPublisher
	ConfigManager  iface.ConfigManager
	NetworkGateway iface.NetworkGateway
}

// depsFromRuntime 从 Runtime 组装 Bag 所需依赖
func depsFromRuntime(rt *runtime.Runtime) Deps {
	return Deps{
		PlayerRepo:     rt.PlayerRepo(),
		EventPublisher: rt.EventPublisher(),
		ConfigManager:  rt.ConfigManager(),
		NetworkGateway: rt.NetworkGateway(),
	}
}
