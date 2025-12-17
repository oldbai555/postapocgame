package deps

import (
	"postapocgame/server/service/gameserver/internel/app/manager"
	"postapocgame/server/service/gameserver/internel/app/playeractor/event"
	gateway2 "postapocgame/server/service/gameserver/internel/app/playeractor/gateway"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Phase 2D 后：移除全局单例，只提供工厂函数
// 约束：
//   1）绝大部分业务代码（SystemAdapter / Service / UseCase / Presenter / Controller）不得直接调用 NewXXX，
//      必须通过 Runtime 或各自的 Deps 结构注入依赖。
//   2）仅以下场景允许使用：
//        - runtime.NewRuntime 内部组装依赖；
//        - 极少量 bootstrapping 代码（如协议路由器/账号控制器的构造函数、兼容性适配器 bag.NewBagUseCaseAdapter）；
//      其他新代码如需调用 NewXXX，必须在评审时给出明确理由。

// NewPlayerGateway 创建 PlayerRepository 实例
func NewPlayerGateway() iface.PlayerRepository {
	return gateway2.NewPlayerGateway()
}

// NewAccountRepository 创建 AccountRepository 实例
func NewAccountRepository() iface.AccountRepository {
	return gateway2.NewAccountGateway()
}

// NewRoleRepository 创建 RoleRepository 实例
func NewRoleRepository() iface.RoleRepository {
	return gateway2.NewRoleGateway()
}

// NewNetworkGateway 创建 ClientGateway 实例（包含 NetworkGateway 和 SessionGateway）
func NewNetworkGateway() gateway2.ClientGateway {
	return gateway2.NewClientGateway()
}

// NewPublicActorGateway 创建 PublicActorGateway 实例
func NewPublicActorGateway() iface.PublicActorGateway {
	return gateway2.NewPublicActorGateway()
}

// NewDungeonServerGateway 创建 DungeonServerGateway 实例
func NewDungeonServerGateway() iface.DungeonServerGateway {
	return gateway2.NewDungeonServerGateway()
}

// NewConfigManager 创建 ConfigManager 实例
func NewConfigManager() iface.ConfigManager {
	return gateway2.NewConfigGateway()
}

// NewEventPublisher 创建 EventPublisher 实例
func NewEventPublisher() iface.EventPublisher {
	return event.NewEventAdapter()
}

// NewTokenGenerator 创建 TokenGenerator 实例
func NewTokenGenerator() iface.TokenGenerator {
	return gateway2.NewTokenGenerator()
}

// GetPlayerRoleManager 获取 PlayerRoleManager 单例
func GetPlayerRoleManager() iface.IPlayerRoleManager {
	return manager.GetPlayerRoleManager()
}
