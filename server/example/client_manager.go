package main

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/log"
	"sync"
)

const GatewayAddr = "0.0.0.0:1011"

// ClientManager 客户端管理器（使用Actor模式）
type ClientManager struct {
	actorMgr actor.IActorManager
	clients  map[string]*GameClient
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewClientManager(ctx context.Context) *ClientManager {
	clientCtx, cancel := context.WithCancel(ctx)

	mgr := &ClientManager{
		clients: make(map[string]*GameClient),
		ctx:     clientCtx,
		cancel:  cancel,
	}

	// 创建Actor管理器 (每个客户端一个Actor)
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

// CreateClient 创建客户端
func (cm *ClientManager) CreateClient(playerID string, gatewayAddr string) *GameClient {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client := NewGameClient(playerID, gatewayAddr, cm.actorMgr)
	cm.clients[playerID] = client

	log.Infof("创建客户端: %s", playerID)
	return client
}

// GetClient 获取客户端
func (cm *ClientManager) GetClient(playerID string) *GameClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[playerID]
}

// DestroyClient 关闭并移除指定客户端
func (cm *ClientManager) DestroyClient(playerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if client, ok := cm.clients[playerID]; ok {
		client.Close()
		delete(cm.clients, playerID)
	}
}

// Stop 停止管理器
func (cm *ClientManager) Stop() {
	log.Infof("停止客户端管理器...")

	// 停止所有客户端
	cm.mu.RLock()
	clients := make([]*GameClient, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	cm.mu.RUnlock()

	for _, client := range clients {
		client.Close()
	}

	// 停止Actor管理器
	if err := cm.actorMgr.Stop(cm.ctx); err != nil {
		log.Errorf("stop actor manager failed: %v", err)
	}

	cm.cancel()
	log.Infof("客户端管理器已停止")
}
