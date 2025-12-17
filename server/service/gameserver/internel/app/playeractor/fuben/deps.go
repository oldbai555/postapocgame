package fuben

import (
	"postapocgame/server/service/gameserver/internel/app/playeractor/level"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/app/playeractor/service/consume"
	"postapocgame/server/service/gameserver/internel/app/playeractor/service/reward"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Deps 聚合 Fuben 相关依赖
type Deps struct {
	PlayerRepo     iface.PlayerRepository
	ConfigManager  iface.ConfigManager
	EventPublisher iface.EventPublisher
	DungeonGateway iface.DungeonServerGateway
	NetworkGateway iface.NetworkGateway
	ConsumeUseCase iface.ConsumeUseCase
	LevelUseCase   iface.LevelUseCase
	RewardUseCase  iface.RewardUseCase
}

// depsFromRuntime 从 Runtime 组装 Fuben 所需依赖
func depsFromRuntime(rt *runtime.Runtime) Deps {
	return Deps{
		PlayerRepo:     rt.PlayerRepo(),
		ConfigManager:  rt.ConfigManager(),
		EventPublisher: rt.EventPublisher(),
		DungeonGateway: rt.DungeonGateway(),
		NetworkGateway: rt.NetworkGateway(),
		ConsumeUseCase: consume.NewConsumeUseCase(rt.PlayerRepo(), rt.EventPublisher()), // TODO: 后续可考虑挂到 Runtime
		LevelUseCase:   level.NewLevelUseCaseAdapter(),
		RewardUseCase:  reward.NewRewardUseCase(rt.PlayerRepo(), rt.EventPublisher(), rt.ConfigManager()),
	}
}
