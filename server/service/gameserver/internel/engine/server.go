package engine

import (
	"context"
	"fmt"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/config"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"sync"
)

// GameServer GameServer实现
type GameServer struct {
	config    *config.ServerConfig
	tcpServer network.ITCPServer

	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewGameServer 创建GameServer
func NewGameServer(config *config.ServerConfig) *GameServer {
	return &GameServer{
		config:   config,
		stopChan: make(chan struct{}),
	}
}

func (gs *GameServer) Start(ctx context.Context) error {
	log.Infof("Starting GameServer: AppID=%d, PlatformId=%d, SrvId=%d", gs.config.AppID, gs.config.PlatformID, gs.config.SrvId)
	// 启动TCP服务器
	serverConfig := &network.TCPServerConfig{
		Addr:            gs.config.TCPAddr,
		AllowedIPs:      gs.config.GatewayAllowIPs,
		MaxConnections:  10,
		HandshakeEnable: false,
	}
	networkHandler := gatewaylink.DefaultNetworkHandler()
	gs.tcpServer = network.NewTCPServer(serverConfig, networkHandler)
	if err := gs.tcpServer.Start(ctx); err != nil {
		return fmt.Errorf("start tcp service failed: %w", err)
	}
	log.Infof("GameServer started successfully")
	return nil
}

// Stop 停止服务器
func (gs *GameServer) Stop(ctx context.Context) error {
	close(gs.stopChan)

	// 停止TCP服务器
	if gs.tcpServer != nil {
		gs.tcpServer.Stop(ctx)
	}

	gs.wg.Wait()
	log.Infof("GameServer stopped")
	return nil
}

func (gs *GameServer) GetConfig() *config.ServerConfig {
	return gs.config
}
