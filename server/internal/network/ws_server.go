package network

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"sync"
	"time"
)

// WSServerConfig WebSocket服务器配置
type WSServerConfig struct {
	Addr            string                     // 监听地址
	Path            string                     // WebSocket路径
	AllowedIPs      []string                   // 允许的IP列表(为空则允许所有)
	MaxConnections  int                        // 最大连接数
	HandshakeEnable bool                       // 是否启用握手
	CheckOrigin     func(r *http.Request) bool // Origin检查函数
}

// WSServer WebSocket服务器
type WSServer struct {
	config      *WSServerConfig
	handler     INetworkMessageHandler
	httpServer  *http.Server
	connections map[*websocket.Conn]IConnection
	mu          sync.RWMutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
	upgrader    websocket.Upgrader
}

// NewWSServer 创建WebSocket服务器
func NewWSServer(config *WSServerConfig, handler INetworkMessageHandler) *WSServer {
	checkOrigin := config.CheckOrigin
	if checkOrigin == nil {
		checkOrigin = func(r *http.Request) bool { return true }
	}

	return &WSServer{
		config:      config,
		handler:     handler,
		connections: make(map[*websocket.Conn]IConnection),
		stopChan:    make(chan struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: checkOrigin,
		},
	}
}

// Start 启动服务器
func (s *WSServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.config.Path, s.handleWebSocket)

	s.httpServer = &http.Server{
		Addr:    s.config.Addr,
		Handler: mux,
	}

	log.Infof("WebSocket server starting on %s%s", s.config.Addr, s.config.Path)

	s.wg.Add(1)
	routine.GoV2(func() error {
		defer s.wg.Done()
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("WebSocket server error: %v", err)
			return err
		}
		return nil
	})

	return nil
}

// Stop 停止服务器
func (s *WSServer) Stop(ctx context.Context) error {
	close(s.stopChan)

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Errorf("WebSocket server shutdown error: %v", err)
		}
	}

	// 关闭所有连接
	s.mu.Lock()
	for _, conn := range s.connections {
		conn.Close()
	}
	s.mu.Unlock()

	s.wg.Wait()

	log.Infof("WebSocket server stopped")
	return nil
}

// GetAddr 获取监听地址
func (s *WSServer) GetAddr() string {
	return s.config.Addr
}

// handleWebSocket 处理WebSocket连接
func (s *WSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 检查是否应该停止接受新连接
	select {
	case <-s.stopChan:
		http.Error(w, "Service shutting down", http.StatusServiceUnavailable)
		return
	default:
	}

	// 检查IP白名单
	if !s.isIPAllowed(r.RemoteAddr) {
		log.Infof("WebSocket connection rejected: %s", r.RemoteAddr)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 检查连接数限制
	s.mu.RLock()
	connCount := len(s.connections)
	s.mu.RUnlock()

	if s.config.MaxConnections > 0 && connCount >= s.config.MaxConnections {
		log.Infof("Max connections reached, reject: %s", r.RemoteAddr)
		http.Error(w, "Too Many Connections", http.StatusServiceUnavailable)
		return
	}

	// 升级到WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Upgrade to WebSocket failed: %v", err)
		return
	}

	// 创建连接对象
	wsConn := NewWebSocketConnection(conn)

	s.mu.Lock()
	s.connections[conn] = wsConn
	s.mu.Unlock()

	log.Infof("New WebSocket connection: %s", conn.RemoteAddr().String())

	// 启动连接处理协程
	s.wg.Add(1)
	routine.GoV2(func() error {
		s.handleConnection(context.Background(), wsConn, conn)
		return nil
	})
}

// handleConnection 处理单个连接
func (s *WSServer) handleConnection(ctx context.Context, wsConn IConnection, rawConn *websocket.Conn) {
	defer s.wg.Done()
	defer func() {
		wsConn.Close()
		s.mu.Lock()
		delete(s.connections, rawConn)
		s.mu.Unlock()
		log.Infof("WebSocket connection closed: %s", rawConn.RemoteAddr().String())
	}()

	const defaultHeartbeatTimeout = 60 * time.Second
	lastActive := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
		}

		if time.Since(lastActive) > defaultHeartbeatTimeout {
			log.Warnf("[HEARTBEAT] ws connection idle timeout from %s, kicking...", rawConn.RemoteAddr().String())
			return
		}

		msg, err := wsConn.ReceiveMessage(ctx)
		if err != nil {
			log.Errorf("Receive message failed: %v", err)
			return
		}
		lastActive = time.Now()
		if msg.Type == MsgTypeHeartbeat {
			log.Debugf("[HEARTBEAT] recv hb from %s", rawConn.RemoteAddr().String())
			continue
		}
		// 调用消息处理器
		if err := s.handler.HandleMessage(ctx, wsConn, msg); err != nil {
			log.Errorf("Handle message failed: %v", err)
		}
	}
}

// handleHandshake 处理握手
func (s *WSServer) handleHandshake(ctx context.Context, conn IConnection) error {
	msg, err := conn.ReceiveMessage(ctx)
	if err != nil {
		return fmt.Errorf("receive handshake failed: %w", err)
	}

	if msg.Type != MsgTypeHandshake {
		return fmt.Errorf("expected handshake message, got %d", msg.Type)
	}

	codec := DefaultCodec()
	handshake, err := codec.DecodeHandshake(msg.Payload)
	if err != nil {
		return fmt.Errorf("decode handshake failed: %w", err)
	}

	// 保存握手信息到连接元数据
	conn.SetMeta(handshake)

	log.Infof("Handshake success: ServerType=%d, PlatformId=%d, ZoneId=%d, SrvType=%d",
		handshake.ServerType, handshake.PlatformId, handshake.ZoneId, handshake.SrvType)

	return nil
}

// isIPAllowed 检查IP是否允许
func (s *WSServer) isIPAllowed(addr string) bool {
	if len(s.config.AllowedIPs) == 0 {
		return true
	}

	// 提取IP地址
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}

	// 检查白名单
	for _, allowedIP := range s.config.AllowedIPs {
		if host == allowedIP || allowedIP == "0.0.0.0" {
			return true
		}
	}

	return false
}

// GetConnections 获取所有连接
func (s *WSServer) GetConnections() []IConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns := make([]IConnection, 0, len(s.connections))
	for _, conn := range s.connections {
		conns = append(conns, conn)
	}
	return conns
}

// GetConnectionCount 获取连接数
func (s *WSServer) GetConnectionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections)
}
