package di

import (
	"postapocgame/server/service/gameserver/internel/adapter/event"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	// Gateways
	playerGateway        repository.PlayerRepository
	networkGateway       gateway.NetworkGateway
	publicActorGateway   interfaces.PublicActorGateway
	dungeonServerGateway interfaces.DungeonServerGateway
	configGateway        interfaces.ConfigManager
	eventPublisher       interfaces.EventPublisher
	blacklistRepository  interfaces.BlacklistRepository

	// Use Cases（按依赖顺序初始化）
	// 注意：Use Cases 将在后续阶段添加

	// Controllers（按依赖顺序初始化）
	// 注意：Controllers 将在后续阶段添加

	// Presenters（按依赖顺序初始化）
	// 注意：Presenters 将在后续阶段添加
}

var (
	globalContainer *Container
	containerOnce   sync.Once
)

// GetContainer 获取全局依赖注入容器
func GetContainer() *Container {
	containerOnce.Do(func() {
		globalContainer = NewContainer()
	})
	return globalContainer
}

// NewContainer 创建依赖注入容器
func NewContainer() *Container {
	c := &Container{}

	// 初始化 Gateways（无依赖，最先初始化）
	c.playerGateway = gateway.NewPlayerGateway()
	c.networkGateway = gateway.NewNetworkGateway()
	c.publicActorGateway = gateway.NewPublicActorGateway()
	c.dungeonServerGateway = gateway.NewDungeonServerGateway()
	c.configGateway = gateway.NewConfigGateway()
	c.eventPublisher = event.NewEventAdapter()
	c.blacklistRepository = gateway.NewBlacklistRepositoryAdapter()

	// Use Cases 将在后续阶段添加
	// 例如：
	// c.levelUseCase = level.NewLevelUseCase(c.playerGateway, c.eventPublisher)

	// Controllers 将在后续阶段添加
	// 例如：
	// c.levelController = controller.NewLevelController(c.levelUseCase, ...)

	// Presenters 将在后续阶段添加
	// 例如：
	// c.levelPresenter = presenter.NewLevelPresenter(c.networkGateway)

	return c
}

// Getter 方法（供外部访问）
func (c *Container) GetPlayerGateway() repository.PlayerRepository {
	return c.playerGateway
}

func (c *Container) GetNetworkGateway() gateway.NetworkGateway {
	return c.networkGateway
}

func (c *Container) GetPublicActorGateway() interfaces.PublicActorGateway {
	return c.publicActorGateway
}

func (c *Container) GetDungeonServerGateway() interfaces.DungeonServerGateway {
	return c.dungeonServerGateway
}

func (c *Container) GetConfigGateway() interfaces.ConfigManager {
	return c.configGateway
}

func (c *Container) GetEventPublisher() interfaces.EventPublisher {
	return c.eventPublisher
}

func (c *Container) GetBlacklistRepository() interfaces.BlacklistRepository {
	return c.blacklistRepository
}

// 为了向后兼容，添加公共字段访问（不推荐，但为了简化代码）
func (c *Container) PlayerGateway() repository.PlayerRepository {
	return c.playerGateway
}

func (c *Container) NetworkGateway() gateway.NetworkGateway {
	return c.networkGateway
}

func (c *Container) PublicActorGateway() interfaces.PublicActorGateway {
	return c.publicActorGateway
}

func (c *Container) DungeonServerGateway() interfaces.DungeonServerGateway {
	return c.dungeonServerGateway
}

func (c *Container) ConfigGateway() interfaces.ConfigManager {
	return c.configGateway
}

func (c *Container) EventPublisher() interfaces.EventPublisher {
	return c.eventPublisher
}

func (c *Container) BlacklistRepository() interfaces.BlacklistRepository {
	return c.blacklistRepository
}
