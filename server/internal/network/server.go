package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"sync"
	"sync/atomic"
	"time"
)

// TCPServerConfig TCP服务器配置
type TCPServerConfig struct {
	Addr            string   // 监听地址
	AllowedIPs      []string // 允许的IP列表(为空则允许所有)
	MaxConnections  int      // 最大连接数
	HandshakeEnable bool     // 是否启用握手
}

// TCPServer TCP服务器
type TCPServer struct {
	config      *TCPServerConfig
	handler     INetworkMessageHandler
	listener    net.Listener
	connections sync.Map // map[net.Conn]IConnection
	stopChan    chan struct{}
	stopping    atomic.Bool
	wg          sync.WaitGroup

	connCount atomic.Int32
	closeOnce sync.Once
}

// NewTCPServer 创建TCP服务器
func NewTCPServer(config *TCPServerConfig, handler INetworkMessageHandler) *TCPServer {
	return &TCPServer{
		config:   config,
		handler:  handler,
		stopChan: make(chan struct{}),
	}
}

// Start 启动服务器
func (s *TCPServer) Start(ctx context.Context) error {
	if s.handler == nil {
		return fmt.Errorf("no message handler provided")
	}

	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}

	s.listener = listener
	log.Infof("✅ TCP service started on %s", s.config.Addr)

	// 启动接受连接协程
	s.wg.Add(1)
	routine.GoV2(func() error {
		s.acceptLoop(ctx)
		return nil
	})

	return nil
}

// Stop 停止服务器
func (s *TCPServer) Stop(ctx context.Context) error {
	s.closeOnce.Do(func() {
		s.stopping.Store(true)
		close(s.stopChan)

		if s.listener != nil {
			s.listener.Close()
		}

		// 并发关闭所有连接
		s.connections.Range(func(key, value any) bool {
			if conn, ok := value.(IConnection); ok {
				conn.Close()
			}
			return true
		})

		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Infof("✅ TCP service stopped (%d connections closed)", s.connCount.Load())
		case <-time.After(10 * time.Second):
			log.Warnf("⏳ timeout waiting for server shutdown")
		}
	})
	return nil
}

// acceptLoop 接受连接循环
func (s *TCPServer) acceptLoop(ctx context.Context) {
	defer s.wg.Done()
	defer log.Infof("accept loop exited")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			select {
			case <-s.stopChan:
				return
			default:
				log.Warnf("accept connection failed: %v", err)
				time.Sleep(200 * time.Millisecond) // 小延迟防止 busy loop
				continue
			}
		}

		remote := conn.RemoteAddr().String()

		// 检查IP白名单
		if !s.isIPAllowed(conn.RemoteAddr()) {
			log.Warnf("🚫 connection rejected (IP not allowed): %s", remote)
			conn.Close()
			continue
		}

		// 检查连接数限制
		if s.config.MaxConnections > 0 && int(s.connCount.Load()) >= s.config.MaxConnections {
			log.Warnf("🚫 max connections reached, reject: %s", remote)
			conn.Close()
			continue
		}

		tcpConn := NewTCPConnection(conn)
		s.connections.Store(conn, tcpConn)
		s.connCount.Add(1)

		log.Infof("🔌 new connection: %s (total=%d)", remote, s.connCount.Load())

		// 启动连接处理协程
		s.wg.Add(1)
		routine.GoV2(func() error {
			s.handleConnection(ctx, tcpConn, conn)
			return nil
		})
	}
}

// handleConnection 处理单个连接
func (s *TCPServer) handleConnection(ctx context.Context, tcpConn IConnection, rawConn net.Conn) {
	defer s.wg.Done()
	defer func() {
		tcpConn.Close()
		s.connections.Delete(rawConn)
		s.connCount.Add(-1)
		log.Infof("❌ connection closed: %s (remaining=%d)", rawConn.RemoteAddr().String(), s.connCount.Load())
	}()

	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 握手阶段
	if s.config.HandshakeEnable {
		if err := s.handleHandshake(connCtx, tcpConn); err != nil {
			log.Warnf("handshake failed from %s: %v", rawConn.RemoteAddr().String(), err)
			return
		}
		log.Infof("🤝 handshake success from %s", rawConn.RemoteAddr().String())
	}

	// 主循环
	for {
		select {
		case <-connCtx.Done():
			return
		case <-s.stopChan:
			return
		default:
		}

		msg, err := tcpConn.ReceiveMessage(connCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed) {
				return
			}
			log.Warnf("receive message failed from %s: %v", rawConn.RemoteAddr().String(), err)
			return
		}

		if err := s.handler.HandleMessage(connCtx, tcpConn, msg); err != nil {
			log.Warnf("handle message failed from %s: %v", rawConn.RemoteAddr().String(), err)
		}
	}
}

// handleHandshake 处理握手
func (s *TCPServer) handleHandshake(ctx context.Context, conn IConnection) error {
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

	conn.SetMeta(handshake)

	log.Infof("handshake success: ServerType=%d, PlatformId=%d, ZoneId=%d, SrvType=%d",
		handshake.ServerType, handshake.PlatformId, handshake.ZoneId, handshake.SrvType)

	return nil
}

// isIPAllowed 检查IP是否允许
func (s *TCPServer) isIPAllowed(addr net.Addr) bool {
	if len(s.config.AllowedIPs) == 0 {
		return true
	}

	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return false
	}
	ip := tcpAddr.IP.String()

	for _, allowed := range s.config.AllowedIPs {
		if ip == allowed || allowed == "0.0.0.0" {
			return true
		}
	}
	return false
}

// GetConnections 获取所有连接
func (s *TCPServer) GetConnections() []IConnection {
	var conns []IConnection
	s.connections.Range(func(_, value any) bool {
		if conn, ok := value.(IConnection); ok {
			conns = append(conns, conn)
		}
		return true
	})
	return conns
}

// GetConnectionCount 获取连接数
func (s *TCPServer) GetConnectionCount() int {
	return int(s.connCount.Load())
}
