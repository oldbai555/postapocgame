package client

import (
	"context"
	"sync"

	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/log"
)

const DefaultGatewayAddr = "0.0.0.0:1011"

// Manager 管理调试客户端的生命周期
type Manager struct {
	actorMgr actor.IActorManager
	clients  map[string]*Core
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager(ctx context.Context) *Manager {
	clientCtx, cancel := context.WithCancel(ctx)

	mgr := &Manager{
		clients: make(map[string]*Core),
		ctx:     clientCtx,
		cancel:  cancel,
	}

	mgr.actorMgr = actor.NewActorManager(
		actor.ModePerKey,
		1000,
		func() actor.IActorHandler {
			return NewClientHandler()
		},
	)

	if err := mgr.actorMgr.Init(); err != nil {
		log.Fatalf("init actor manager failed: %v", err)
	}

	if err := mgr.actorMgr.Start(clientCtx); err != nil {
		log.Fatalf("start actor manager failed: %v", err)
	}

	return mgr
}

// CreateClient 创建新的调试客户端核心
func (cm *Manager) CreateClient(playerID string, gatewayAddr string) *Core {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	core := NewCore(playerID, gatewayAddr, cm.actorMgr)
	cm.clients[playerID] = core

	log.Infof("创建客户端: %s", playerID)
	return core
}

// GetClient 根据 playerID 返回客户端
func (cm *Manager) GetClient(playerID string) *Core {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[playerID]
}

// DestroyClient 关闭指定客户端
func (cm *Manager) DestroyClient(playerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if core, ok := cm.clients[playerID]; ok {
		core.Close()
		delete(cm.clients, playerID)
	}
}

// Stop 停止所有客户端与 Actor 管理器
func (cm *Manager) Stop() {
	log.Infof("停止客户端管理器...")

	cm.mu.RLock()
	clients := make([]*Core, 0, len(cm.clients))
	for _, core := range cm.clients {
		clients = append(clients, core)
	}
	cm.mu.RUnlock()

	for _, core := range clients {
		core.Close()
	}

	if err := cm.actorMgr.Stop(cm.ctx); err != nil {
		log.Errorf("stop actor manager failed: %v", err)
	}

	cm.cancel()
	log.Infof("客户端管理器已停止")
}
