package engine

import (
	"context"
	"fmt"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/config"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"sync"
	"time"
)

// DungeonServer DungeonServer实现
type DungeonServer struct {
	config         *config.ServerConfig
	tcpServer      network.ITCPServer
	networkHandler *gameserverlink.NetworkHandler

	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewDungeonServer 创建DungeonServer
func NewDungeonServer(config *config.ServerConfig) *DungeonServer {
	ds := &DungeonServer{
		config:   config,
		stopChan: make(chan struct{}),
	}
	return ds
}

// Start 启动服务器
func (ds *DungeonServer) Start(ctx context.Context) error {
	log.Infof("Starting DungeonServer: SrvType=%d", ds.config.SrvType)
	// 启动TCP服务器
	if err := ds.startTCPServer(ctx); err != nil {
		return err
	}
	log.Infof("DungeonServer started successfully")
	return nil
}

// startTCPServer 启动TCP服务器
func (ds *DungeonServer) startTCPServer(ctx context.Context) error {
	serverConfig := &network.TCPServerConfig{
		Addr:            ds.config.TCPAddr,
		AllowedIPs:      nil,
		MaxConnections:  100,
		HandshakeEnable: true,
	}

	ds.networkHandler = gameserverlink.NewNetworkHandler()
	ds.tcpServer = network.NewTCPServer(serverConfig, ds.networkHandler)

	if err := ds.tcpServer.Start(ctx); err != nil {
		return fmt.Errorf("start tcp service failed: %w", err)
	}

	return nil
}

// Stop 停止服务器
func (ds *DungeonServer) Stop(ctx context.Context) error {
	log.Infof("DungeonServer shutting down...")

	// 1. 停止接收新连接
	close(ds.stopChan)

	// 2. 刷新所有批处理消息
	if ds.networkHandler != nil {
		ds.networkHandler.Close()
	}

	// 3. 停止TCP服务器
	if ds.tcpServer != nil {
		ds.tcpServer.Stop(ctx)
	}

	// 4. 等待所有请求完成（带超时）
	done := make(chan struct{})
	go func() {
		ds.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Infof("All requests completed gracefully")
	case <-time.After(30 * time.Second):
		log.Warnf("Force shutdown after 30s timeout")
	}

	log.Infof("DungeonServer stopped")
	return nil
}
