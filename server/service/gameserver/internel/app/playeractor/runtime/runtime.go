package runtime

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Runtime 聚合 PlayerActor 运行时依赖，用于替代全局 deps。
// 每个 PlayerRole 持有一个 Runtime 实例，便于测试和依赖注入。
type Runtime struct {
	playerRepo     iface.PlayerRepository
	roleRepo       iface.RoleRepository
	configManager  iface.ConfigManager
	eventPublisher iface.EventPublisher
	networkGateway iface.NetworkGateway
	dungeonGateway iface.DungeonServerGateway
	publicGateway  iface.PublicActorGateway
}

// NewRuntime 创建 Runtime 实例（通过依赖注入）
func NewRuntime(
	playerRepo iface.PlayerRepository,
	roleRepo iface.RoleRepository,
	configManager iface.ConfigManager,
	eventPublisher iface.EventPublisher,
	networkGateway iface.NetworkGateway,
	dungeonGateway iface.DungeonServerGateway,
	publicGateway iface.PublicActorGateway,
) *Runtime {
	return &Runtime{
		playerRepo:     playerRepo,
		roleRepo:       roleRepo,
		configManager:  configManager,
		eventPublisher: eventPublisher,
		networkGateway: networkGateway,
		dungeonGateway: dungeonGateway,
		publicGateway:  publicGateway,
	}
}

// PlayerRepo 获取玩家仓储
func (r *Runtime) PlayerRepo() iface.PlayerRepository {
	return r.playerRepo
}

// RoleRepo 获取角色仓储
func (r *Runtime) RoleRepo() iface.RoleRepository {
	return r.roleRepo
}

// ConfigManager 获取配置管理器
func (r *Runtime) ConfigManager() iface.ConfigManager {
	return r.configManager
}

// EventPublisher 获取事件发布器
func (r *Runtime) EventPublisher() iface.EventPublisher {
	return r.eventPublisher
}

// NetworkGateway 获取网络网关
func (r *Runtime) NetworkGateway() iface.NetworkGateway {
	return r.networkGateway
}

// DungeonGateway 获取副本网关
func (r *Runtime) DungeonGateway() iface.DungeonServerGateway {
	return r.dungeonGateway
}

// PublicGateway 获取公共 Actor 网关
func (r *Runtime) PublicGateway() iface.PublicActorGateway {
	return r.publicGateway
}

// WithContext 将 Runtime 注入到 Context（供需要从 Context 获取 Runtime 的场景使用）
func (r *Runtime) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, runtimeKey, r)
}

// contextKey 类型，避免与其他包的 key 冲突
type contextKey int

const runtimeKey contextKey = 0

// FromContext 从 Context 中获取 Runtime（如果没有则返回 nil）
func FromContext(ctx context.Context) *Runtime {
	if rt, ok := ctx.Value(runtimeKey).(*Runtime); ok {
		return rt
	}
	return nil
}
