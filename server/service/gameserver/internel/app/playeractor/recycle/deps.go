package recycle

import (
	"postapocgame/server/service/gameserver/internel/app/playeractor/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/app/playeractor/service/reward"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Deps 聚合 Recycle 相关依赖
type Deps struct {
	ConfigManager  iface.ConfigManager
	NetworkGateway iface.NetworkGateway
	BagUseCase     iface.BagUseCase
	RewardUseCase  iface.RewardUseCase
}

// depsFromRuntime 从 Runtime 组装 Recycle 所需依赖
func depsFromRuntime(rt *runtime.Runtime) Deps {
	return Deps{
		ConfigManager:  rt.ConfigManager(),
		NetworkGateway: rt.NetworkGateway(),
		BagUseCase:     bag.NewBagUseCaseAdapter(), // TODO: 后续可考虑挂到 Runtime
		RewardUseCase:  reward.NewRewardUseCase(rt.PlayerRepo(), rt.EventPublisher(), rt.ConfigManager()),
	}
}
