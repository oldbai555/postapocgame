package equip

import (
	"postapocgame/server/service/gameserver/internel/app/playeractor/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Deps 聚合 Equip 功能所需的所有外部依赖
type Deps struct {
	PlayerRepo     iface.PlayerRepository
	EventPublisher iface.EventPublisher
	ConfigManager  iface.ConfigManager
	BagUseCase     iface.BagUseCase
	NetworkGateway iface.NetworkGateway
}

// depsFromRuntime 从 Runtime 组装 Equip 所需依赖
func depsFromRuntime(rt *runtime.Runtime) Deps {
	return Deps{
		PlayerRepo:     rt.PlayerRepo(),
		EventPublisher: rt.EventPublisher(),
		ConfigManager:  rt.ConfigManager(),
		NetworkGateway: rt.NetworkGateway(),
		BagUseCase:     bag.NewBagUseCaseAdapter(), // TODO: 后续可考虑将 BagUseCase 也挂到 Runtime
	}
}
