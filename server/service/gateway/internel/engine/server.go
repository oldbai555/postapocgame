package engine

import (
	"context"
	"fmt"
	"net/http"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"postapocgame/server/service/gateway/internel/clientnet"
	"postapocgame/server/service/gateway/internel/gameserverlink"
	"sync"
	"time"
)

type GatewayServer struct {
	config      *Config
	sessionMgr  *clientnet.SessionManager
	gsConnector clientnet.IGameServerConnector
	tcpServer   network.ITCPServer
	wsServer    *network.WSServer
	stopChan    chan struct{}
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewGatewayServer(config *Config) (*GatewayServer, error) {
	gsConnector := gameserverlink.NewGameClient(config.GameServerAddr)
	sessionMgr := clientnet.NewSessionManager(config.MaxSessions, config.SessionBufferSize, config.SessionTimeout, gsConnector)

	return &GatewayServer{
		config:      config,
		sessionMgr:  sessionMgr,
		gsConnector: gsConnector,
		stopChan:    make(chan struct{}),
	}, nil
}

// Start 启动网关
func (g *GatewayServer) Start(ctx context.Context) error {
	g.ctx, g.cancel = context.WithCancel(ctx)

	// 连接到GameServer
	if err := g.gsConnector.Connect(g.ctx, g.config.GameServerAddr); err != nil {
		return fmt.Errorf("connect to game server failed: %w", err)
	}

	// 启动会话清理
	g.sessionMgr.StartCleanup(g.ctx)

	// 启动GameServer消息分发
	g.wg.Add(1)
	routine.GoV2(func() error {
		g.dispatchGameServerMessages(g.ctx)
		return nil
	})

	// 启动TCP服务器
	if g.config.TCPAddr != "" {
		if err := g.startTCPServer(g.ctx); err != nil {
			return fmt.Errorf("start tcp server failed: %w", err)
		}
	}

	// 启动WebSocket服务器
	if g.config.WSAddr != "" {
		if err := g.startWSServer(g.ctx); err != nil {
			return fmt.Errorf("start websocket server failed: %w", err)
		}
	}

	log.Infof("GatewayServer started successfully")
	return nil
}

// Stop 停止网关
func (g *GatewayServer) Stop(ctx context.Context) error {
	log.Infof("GatewayServer stopping...")

	if g.cancel != nil {
		g.cancel()
	}

	select {
	case <-g.stopChan:
	default:
		close(g.stopChan)
	}

	// 停止TCP服务器
	if g.tcpServer != nil {
		log.Infof("Stopping TCP server...")
		g.tcpServer.Stop(ctx)
	}

	// 停止WebSocket服务器
	if g.wsServer != nil {
		log.Infof("Stopping WebSocket server...")
		g.wsServer.Stop(ctx)
	}

	// 停止会话管理器
	log.Infof("Stopping session manager...")
	g.sessionMgr.Stop()

	// 关闭GameServer连接
	log.Infof("Closing game server connector...")
	g.gsConnector.Close()

	// 等待所有协程退出
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Infof("All goroutines exited")
	case <-time.After(10 * time.Second):
		log.Warnf("Timeout waiting for goroutines to exit")
	case <-ctx.Done():
		log.Warnf("Context cancelled while waiting for goroutines")
	}

	log.Infof("GatewayServer stopped")
	return nil
}

// startTCPServer 启动TCP服务器（使用统一接口）
func (g *GatewayServer) startTCPServer(ctx context.Context) error {
	handler := clientnet.NewClientHandler(g.gsConnector, g.sessionMgr)

	g.tcpServer = network.NewTCPServer(
		network.WithTCPServerOptionNetworkMessageHandler(handler),
		network.WithTCPServerOptionOnConn(func(conn network.IConnection) {
			log.Infof("new conn......")
		}),
		network.WithTCPServerOptionOnDisConn(func(conn network.IConnection) {
			log.Infof("dis conn......")
		}),
		network.WithTCPServerOptionMaxConnections(g.config.MaxSessions),
		network.WithTCPServerOptionAddr(g.config.TCPAddr),
	)
	if err := g.tcpServer.Start(ctx); err != nil {
		return err
	}

	log.Infof("TCP server started on %s", g.config.TCPAddr)
	return nil
}

// startWSServer 启动WebSocket服务器（使用统一接口）
func (g *GatewayServer) startWSServer(ctx context.Context) error {
	serverConfig := &network.WSServerConfig{
		Addr:            g.config.WSAddr,
		Path:            g.config.WSPath,
		AllowedIPs:      nil,
		MaxConnections:  g.config.MaxSessions,
		HandshakeEnable: false,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	handler := &clientnet.ClientHandler{
		SessionMgr:  g.sessionMgr,
		GsConnector: g.gsConnector,
		Sessions:    make(map[network.IConnection]*clientnet.Session),
	}

	g.wsServer = network.NewWSServer(serverConfig, handler)
	if err := g.wsServer.Start(ctx); err != nil {
		return err
	}

	log.Infof("WebSocket server started on %s%s", g.config.WSAddr, g.config.WSPath)
	return nil
}

// dispatchGameServerMessages 分发GameServer消息到客户端
func (g *GatewayServer) dispatchGameServerMessages(ctx context.Context) {
	defer g.wg.Done()
	defer log.Infof("dispatchGameServerMessages goroutine exited")

	for {
		receiveCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		msg, err := g.gsConnector.ReceiveGsMessage(receiveCtx)
		cancel()

		releaseMsg := func() {
			if msg != nil {
				network.PutForwardMessage(msg)
				msg = nil
			}
		}

		if err != nil {
			select {
			case <-g.stopChan:
				log.Infof("Dispatch stopped by stopChan")
				releaseMsg()
				return
			case <-ctx.Done():
				log.Infof("Dispatch stopped by context")
				releaseMsg()
				return
			default:
				if err == context.DeadlineExceeded {
					releaseMsg()
					continue
				}
				log.Errorf("Receive message from game server failed: %v", err)
				releaseMsg()
				return
			}
		}

		if msg == nil {
			continue
		}

		// 获取会话
		session, ok := g.sessionMgr.GetSession(msg.SessionId)
		if !ok {
			log.Debugf("Session not found: %s", msg.SessionId)
			releaseMsg()
			continue
		}

		// 发送到客户端（非阻塞）
		select {
		case session.SendChan <- msg.Payload:
			releaseMsg()
		case <-time.After(100 * time.Millisecond):
			log.Warnf("Session send channel full or timeout: %s", msg.SessionId)
			releaseMsg()
		case <-g.stopChan:
			releaseMsg()
			return
		case <-ctx.Done():
			releaseMsg()
			return
		}
	}
}

// GetSession 获取会话
func (g *GatewayServer) GetSession(sessionID string) (*clientnet.Session, bool) {
	return g.sessionMgr.GetSession(sessionID)
}

// CloseSession 关闭会话
func (g *GatewayServer) CloseSession(sessionID string) error {
	return g.sessionMgr.CloseSession(sessionID)
}
