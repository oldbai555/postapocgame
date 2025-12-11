package deps

import (
	"postapocgame/server/service/gameserver/internel/app/manager"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/event"
	gateway2 "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/gateway"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"sync"
)

type Dependencies struct {
	playerGateway        repository.PlayerRepository
	accountRepository    repository.AccountRepository
	roleRepository       repository.RoleRepository
	clientGateway        gateway2.ClientGateway
	publicActorGateway   interfaces.PublicActorGateway
	dungeonServerGateway interfaces.DungeonServerGateway
	configGateway        interfaces.ConfigManager
	eventPublisher       interfaces.EventPublisher
	blacklistRepository  interfaces.BlacklistRepository
	tokenGenerator       interfaces.TokenGenerator
	playerRoleManager    interfaces.IPlayerRoleManager
}

// Container 为兼容旧调用保留的别名，后续可直接使用包级函数替换。
type Container Dependencies

var (
	deps     *Dependencies
	depsOnce sync.Once
)

// init 早期初始化，避免在各处重复判断。
func init() {
	Init()
}

// Init 构建依赖实例（只执行一次）。
func Init() {
	depsOnce.Do(func() {
		deps = &Dependencies{
			playerGateway:        gateway2.NewPlayerGateway(),
			accountRepository:    gateway2.NewAccountGateway(),
			roleRepository:       gateway2.NewRoleGateway(),
			clientGateway:        gateway2.NewClientGateway(),
			publicActorGateway:   gateway2.NewPublicActorGateway(),
			dungeonServerGateway: gateway2.NewDungeonServerGateway(),
			configGateway:        gateway2.NewConfigGateway(),
			eventPublisher:       event.NewEventAdapter(),
			blacklistRepository:  gateway2.NewBlacklistRepositoryAdapter(),
			tokenGenerator:       gateway2.NewTokenGenerator(),
			playerRoleManager:    manager.GetPlayerRoleManager(),
		}
	})
}

func ensure() *Dependencies {
	Init()
	if deps == nil {
		panic("deps container not initialized")
	}
	return deps
}

// GetContainer 兼容旧 DI 样式，返回只读视图。
func GetContainer() *Container {
	return (*Container)(ensure())
}

func PlayerGateway() repository.PlayerRepository {
	return ensure().playerGateway
}

func AccountRepository() repository.AccountRepository {
	return ensure().accountRepository
}

func RoleRepository() repository.RoleRepository {
	return ensure().roleRepository
}

func NetworkGateway() gateway2.NetworkGateway { return ensure().clientGateway }
func SessionGateway() gateway2.SessionGateway { return ensure().clientGateway }
func ClientGateway() gateway2.ClientGateway   { return ensure().clientGateway }

func PublicActorGateway() interfaces.PublicActorGateway {
	return ensure().publicActorGateway
}

func DungeonServerGateway() interfaces.DungeonServerGateway {
	return ensure().dungeonServerGateway
}

func ConfigGateway() interfaces.ConfigManager {
	return ensure().configGateway
}

func EventPublisher() interfaces.EventPublisher {
	return ensure().eventPublisher
}

func BlacklistRepository() interfaces.BlacklistRepository {
	return ensure().blacklistRepository
}

func TokenGenerator() interfaces.TokenGenerator {
	return ensure().tokenGenerator
}

func PlayerRoleManager() interfaces.IPlayerRoleManager {
	return ensure().playerRoleManager
}

func (c *Container) PlayerGateway() repository.PlayerRepository { return ensure().playerGateway }
func (c *Container) AccountRepository() repository.AccountRepository {
	return ensure().accountRepository
}
func (c *Container) RoleRepository() repository.RoleRepository { return ensure().roleRepository }
func (c *Container) NetworkGateway() gateway2.NetworkGateway   { return ensure().clientGateway }
func (c *Container) SessionGateway() gateway2.SessionGateway   { return ensure().clientGateway }
func (c *Container) ClientGateway() gateway2.ClientGateway     { return ensure().clientGateway }
func (c *Container) PublicActorGateway() interfaces.PublicActorGateway {
	return ensure().publicActorGateway
}
func (c *Container) DungeonServerGateway() interfaces.DungeonServerGateway {
	return ensure().dungeonServerGateway
}
func (c *Container) ConfigGateway() interfaces.ConfigManager   { return ensure().configGateway }
func (c *Container) EventPublisher() interfaces.EventPublisher { return ensure().eventPublisher }
func (c *Container) BlacklistRepository() interfaces.BlacklistRepository {
	return ensure().blacklistRepository
}
func (c *Container) TokenGenerator() interfaces.TokenGenerator { return ensure().tokenGenerator }
func (c *Container) PlayerRoleManager() interfaces.IPlayerRoleManager {
	return ensure().playerRoleManager
}
