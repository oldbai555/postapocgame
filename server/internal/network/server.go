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

type TCPServerOption func(*TCPServer)

func WithTCPServerOptionOnConn(f func(conn IConnection)) TCPServerOption {
	return func(tcpServer *TCPServer) {
		tcpServer.onConnected = f
	}
}

func WithTCPServerOptionOnDisConn(f func(conn IConnection)) TCPServerOption {
	return func(tcpServer *TCPServer) {
		tcpServer.onDisconnected = f
	}
}

func WithTCPServerOptionNetworkMessageHandler(handler INetworkMessageHandler) TCPServerOption {
	return func(tcpServer *TCPServer) {
		tcpServer.handler = handler
	}
}
func WithTCPServerOptionAllowedIPs(allowedIPs []string) TCPServerOption {
	return func(tcpServer *TCPServer) {
		tcpServer.allowedIPs = allowedIPs
	}
}

func WithTCPServerOptionAddr(addr string) TCPServerOption {
	return func(tcpServer *TCPServer) {
		tcpServer.addr = addr
	}
}
func WithTCPServerOptionMaxConnections(maxConnections uint32) TCPServerOption {
	return func(tcpServer *TCPServer) {
		tcpServer.maxConnections = maxConnections
	}
}

// TCPServer TCP服务器
type TCPServer struct {
	addr           string   // 监听地址
	allowedIPs     []string // 允许的IP列表(为空则允许所有)
	maxConnections uint32   // 最大连接数

	onConnected    func(conn IConnection)
	onDisconnected func(conn IConnection)

	handler INetworkMessageHandler

	listener    net.Listener
	connections sync.Map // map[net.Conn]IConnection
	stopChan    chan struct{}
	stopping    atomic.Bool
	wg          sync.WaitGroup

	connCount atomic.Int32
	closeOnce sync.Once
}

// NewTCPServer 创建TCP服务器
func NewTCPServer(opts ...TCPServerOption) ITCPServer {
	t := &TCPServer{
		stopChan:       make(chan struct{}),
		maxConnections: 10,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Start 启动服务器
func (s *TCPServer) Start(ctx context.Context) error {
	if s.handler == nil {
		return fmt.Errorf("no message handler provided")
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}

	s.listener = listener
	log.Infof("✅ TCP service started on %s", s.addr)

	// 启动接受连接协程
	s.wg.Add(1)
	routine.GoV2(func() error {
		defer func() {
			s.wg.Done()
		}()
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
		if s.maxConnections > 0 && uint32(s.connCount.Load()) >= s.maxConnections {
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
			defer func() {
				s.wg.Done()
			}()
			s.handleConnection(ctx, tcpConn, conn)
			return nil
		})
	}
}

// handleConnection 处理单个连接
func (s *TCPServer) handleConnection(ctx context.Context, tcpConn IConnection, rawConn net.Conn) {
	routine.Run(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("[PANIC] TCP connection handler crashed, remote=%s, err=%v", rawConn.RemoteAddr().String(), r)
			}
			if s.onDisconnected != nil {
				s.onDisconnected(tcpConn)
			}
			tcpConn.Close()
			s.connections.Delete(rawConn)
			s.connCount.Add(-1)
			log.Infof("❌ connection closed: %s (remaining=%d)", rawConn.RemoteAddr().String(), s.connCount.Load())
		}()

		connCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// 握手阶段
		if s.onConnected != nil {
			s.onConnected(tcpConn)
		}

		const defaultHeartbeatTimeout = 60 * time.Second // 心跳或消息的最大空闲时长
		lastActive := time.Now()

		// 主循环
		for {
			select {
			case <-connCtx.Done():
				return
			case <-s.stopChan:
				return
			default:
			}

			// 空闲超时检测
			if time.Since(lastActive) > defaultHeartbeatTimeout {
				log.Warnf("[HEARTBEAT] connection idle timeout from %s, kicking...", rawConn.RemoteAddr().String())
				return
			}

			tcpConnRaw := tcpConn // 保持变量一致，兼容类型
			// 设置读取超时，实现心跳自动检测
			raw, ok := tcpConnRaw.(*TCPConnection)
			if ok {
				_ = raw.conn.SetReadDeadline(time.Now().Add(defaultHeartbeatTimeout))
			}

			msg, err := tcpConn.ReceiveMessage(connCtx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed) {
					return
				}
				// 详细分类日志
				switch {
				case errors.Is(err, ErrFrameTooLarge):
					log.Warnf("[SECURITY] recv frame too large from %s: %v", rawConn.RemoteAddr().String(), err)
				case errors.Is(err, ErrInvalidMessage):
					log.Warnf("[PROTO] invalid message from %s: %v", rawConn.RemoteAddr().String(), err)
				default:
					log.Warnf("receive message failed from %s: %v", rawConn.RemoteAddr().String(), err)
				}
				return
			}

			lastActive = time.Now()
			if msg.Type == MsgTypeHeartbeat {
				continue
			}

			if err := s.handler.HandleMessage(connCtx, tcpConn, msg); err != nil {
				log.Warnf("handle message failed from %s: %v", rawConn.RemoteAddr().String(), err)
			}
		}
	})
}

// isIPAllowed 检查IP是否允许
func (s *TCPServer) isIPAllowed(addr net.Addr) bool {
	if len(s.allowedIPs) == 0 {
		return true
	}

	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return false
	}
	ip := tcpAddr.IP.String()

	for _, allowed := range s.allowedIPs {
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

func (s *TCPServer) SetCallbacks(onConnected, onDisconnected func(conn IConnection)) {
	s.onConnected = onConnected
	s.onDisconnected = onDisconnected
}
