package skill

import (
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Deps 聚合 Skill 功能所需的所有外部依赖
type Deps struct {
	PlayerRepo     iface.PlayerRepository
	ConfigManager  iface.ConfigManager
	EventPublisher iface.EventPublisher
	DungeonGateway iface.DungeonServerGateway
	NetworkGateway iface.NetworkGateway
}

// depsFromRuntime 从 Runtime 组装 Skill 所需依赖
func depsFromRuntime(rt *runtime.Runtime) Deps {
	return Deps{
		PlayerRepo:     rt.PlayerRepo(),
		ConfigManager:  rt.ConfigManager(),
		EventPublisher: rt.EventPublisher(),
		DungeonGateway: rt.DungeonGateway(),
		NetworkGateway: rt.NetworkGateway(),
	}
}
