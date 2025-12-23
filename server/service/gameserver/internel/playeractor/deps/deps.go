package deps

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/gateway"
)

// NewPlayerGateway 创建 PlayerRepository 实例
func NewPlayerGateway() iface.PlayerRepository {
	return gateway.NewPlayerGateway()
}

// NewAccountRepository 创建 AccountRepository 实例
func NewAccountRepository() iface.AccountRepository {
	return gateway.NewAccountGateway()
}

// NewRoleRepository 创建 RoleRepository 实例
func NewRoleRepository() iface.RoleRepository {
	return gateway.NewRoleGateway()
}

// NewNetworkGateway 创建 ClientGateway 实例（包含 NetworkGateway 和 SessionGateway）
func NewNetworkGateway() gateway.ClientGateway {
	return gateway.NewClientGateway()
}

// NewDungeonServerGateway 创建 DungeonServerGateway 实例
func NewDungeonServerGateway() iface.DungeonServerGateway {
	return gateway.NewDungeonServerGateway()
}

// NewTokenGenerator 创建 TokenGenerator 实例
func NewTokenGenerator() iface.TokenGenerator {
	return gateway.NewTokenGenerator()
}

// GetPlayerRoleManager 获取 PlayerRoleManager 单例
func GetPlayerRoleManager() iface.IPlayerRoleManager {
	return manager.GetPlayerRoleManager()
}

// Runtime 聚合 PlayerActor 运行时依赖，用于替代全局 deps。
// 每个 PlayerRole 持有一个 Runtime 实例，便于测试和依赖注入。
type Runtime struct {
	playerRepo     iface.PlayerRepository
	roleRepo       iface.RoleRepository
	networkGateway iface.NetworkGateway
	dungeonGateway iface.DungeonServerGateway
}

// NewRuntime 创建 Runtime 实例（通过依赖注入）
func NewRuntime(
	playerRepo iface.PlayerRepository,
	roleRepo iface.RoleRepository,
	networkGateway iface.NetworkGateway,
	dungeonGateway iface.DungeonServerGateway,
) *Runtime {
	return &Runtime{
		playerRepo:     playerRepo,
		roleRepo:       roleRepo,
		networkGateway: networkGateway,
		dungeonGateway: dungeonGateway,
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

// NetworkGateway 获取网络网关
func (r *Runtime) NetworkGateway() iface.NetworkGateway {
	return r.networkGateway
}

// DungeonGateway 获取副本网关
func (r *Runtime) DungeonGateway() iface.DungeonServerGateway {
	return r.dungeonGateway
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
