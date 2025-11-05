/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package internel

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"sync"
)

// Gateway 网关实现
type Gateway struct {
	config        *Config
	sessionMgr    *SessionManager
	gsConnector   IGameServerConnector
	authenticator IAuthenticator
	tcpListener   net.Listener
	wsServer      *http.Server
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewGateway 创建网关
func NewGateway(config *Config, authenticator IAuthenticator) (*Gateway, error) {
	// 创建GameServer连接器(连接第一个GameServer)
	gsConnector := NewGameServerConnector(config.GameServerAddr, config)

	// 创建会话管理器
	sessionMgr := NewSessionManager(config, gsConnector)

	return &Gateway{
		config:        config,
		sessionMgr:    sessionMgr,
		gsConnector:   gsConnector,
		authenticator: authenticator,
		stopChan:      make(chan struct{}),
	}, nil
}

// Start 启动网关
func (g *Gateway) Start(ctx context.Context) error {
	// 连接到GameServer
	if err := g.gsConnector.Connect(ctx, g.config.GameServerAddr); err != nil {
		return fmt.Errorf("connect to game server failed: %w", err)
	}

	// 启动会话清理
	g.sessionMgr.StartCleanup(ctx)

	// 启动GameServer消息分发
	g.wg.Add(1)
	routine.GoV2(func() error {
		g.dispatchGameServerMessages(ctx)
		return nil
	})

	// 启动TCP监听
	if g.config.TCPAddr != "" {
		if err := g.startTCPListener(ctx); err != nil {
			return fmt.Errorf("start tcp listener failed: %w", err)
		}
	}

	// 启动WebSocket监听
	if g.config.WSAddr != "" {
		if err := g.startWSListener(ctx); err != nil {
			return fmt.Errorf("start websocket listener failed: %w", err)
		}
	}

	fmt.Println("gateway started")
	return nil
}

// Stop 停止网关
func (g *Gateway) Stop(ctx context.Context) error {
	close(g.stopChan)

	// 停止监听器
	if g.tcpListener != nil {
		g.tcpListener.Close()
	}
	if g.wsServer != nil {
		g.wsServer.Shutdown(ctx)
	}

	// 等待所有协程退出
	g.wg.Wait()

	// 停止会话管理器
	g.sessionMgr.Stop()

	// 关闭GameServer连接
	g.gsConnector.Close()

	fmt.Println("gateway stopped")
	return nil
}

// startTCPListener 启动TCP监听
func (g *Gateway) startTCPListener(ctx context.Context) error {
	listener, err := net.Listen("tcp", g.config.TCPAddr)
	if err != nil {
		return err
	}

	g.tcpListener = listener
	log.Infof("tcp listener started on %s", g.config.TCPAddr)

	g.wg.Add(1)
	routine.GoV2(func() error {
		defer g.wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-g.stopChan:
					log.Infof("stop tcp......")
					return nil
				default:
					log.Infof("accept tcp connection failed: %v", err)
					continue
				}
			}

			// 处理新连接
			tcpConn := NewTCPConnection(conn, g.config.MaxFrameSize)
			handler := NewConnectionHandler(
				tcpConn,
				g.sessionMgr,
				g.gsConnector,
				g.authenticator,
				g.config,
			)

			g.wg.Add(1)
			go func() {
				defer g.wg.Done()
				err := handler.Handle(ctx)
				if err != nil {
					log.Errorf("err:%v", err)
				}
			}()
		}
	})

	return nil
}

// startWSListener 启动WebSocket监听
func (g *Gateway) startWSListener(ctx context.Context) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 生产环境需要验证origin
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(g.config.WSPath, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Infof("upgrade websocket failed: %v", err)
			return
		}

		// 处理新连接
		wsConn := NewWebSocketConnection(conn, g.config.MaxFrameSize)
		handler := NewConnectionHandler(
			wsConn,
			g.sessionMgr,
			g.gsConnector,
			g.authenticator,
			g.config,
		)

		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			handler.Handle(ctx)
		}()
	})

	g.wsServer = &http.Server{
		Addr:    g.config.WSAddr,
		Handler: mux,
	}

	log.Infof("websocket listener started on %s%s", g.config.WSAddr, g.config.WSPath)

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := g.wsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Infof("websocket server error: %v", err)
		}
	}()

	return nil
}

// dispatchGameServerMessages 分发GameServer消息到客户端
func (g *Gateway) dispatchGameServerMessages(ctx context.Context) {
	defer g.wg.Done()

	for {
		msg, err := g.gsConnector.ReceiveMessage(ctx)
		if err != nil {
			select {
			case <-g.stopChan:
				return
			default:
				log.Infof("receive message from game server failed: %v", err)
				return
			}
		}

		// 获取会话
		session, ok := g.sessionMgr.GetSession(msg.SessionID)
		if !ok {
			log.Infof("session not found: %s", msg.SessionID)
			continue
		}

		// 发送到客户端
		select {
		case session.SendChan <- msg.Payload:
		default:
			log.Infof("session send channel full: %s", msg.SessionID)
		}
	}
}

// GetSession 获取会话
func (g *Gateway) GetSession(sessionID string) (*Session, bool) {
	return g.sessionMgr.GetSession(sessionID)
}

// CloseSession 关闭会话
func (g *Gateway) CloseSession(sessionID string) error {
	return g.sessionMgr.CloseSession(sessionID)
}

// GetStats 获取统计信息
func (g *Gateway) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"session_count": g.sessionMgr.GetSessionCount(),
	}
}
