package network

import (
	"context"
	"fmt"
	"net"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"sync"
	"sync/atomic"
	"time"

	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
)

type TCPClientOption func(*TCPClient)

func WithTCPClientOptionOnConn(f func(conn IConnection)) TCPClientOption {
	return func(client *TCPClient) {
		client.onConnected = f
	}
}

func WithTCPClientOptionOnDisConn(f func(conn IConnection)) TCPClientOption {
	return func(client *TCPClient) {
		client.onDisconnected = f
	}
}

func WithTCPClientOptionNetworkMessageHandler(handler INetworkMessageHandler) TCPClientOption {
	return func(client *TCPClient) {
		client.handler = handler
	}
}

// TCPClient TCP客户端
type TCPClient struct {
	addr string

	ConnectTimeout  time.Duration // 连接超时
	EnableReconnect bool          // 是否启用重连
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	MaxRetries      int

	conn      IConnection
	connected atomic.Bool
	stopping  atomic.Bool

	reconnecting atomic.Bool
	retryCount   atomic.Int32

	onConnected    func(conn IConnection)
	onDisconnected func(conn IConnection)

	handler INetworkMessageHandler

	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup

	receiveCtx    context.Context
	receiveCancel context.CancelFunc
	receiveMu     sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

func NewTCPClient(opts ...TCPClientOption) ITCPClient {
	t := &TCPClient{
		EnableReconnect: true,
		ConnectTimeout:  3 * time.Second,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      1.5,
		stopChan:        make(chan struct{}),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (c *TCPClient) Connect(ctx context.Context, addr string) error {
	if c.handler == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not found network message handler")
	}
	c.addr = addr

	// 🔹内部 ctx 派生自外部 ctx，外部 ctx 取消 => 客户端关闭
	c.ctx, c.cancel = context.WithCancel(ctx)

	if c.EnableReconnect {
		return c.connectWithReconnect(ctx)
	}
	return c.doConnect(ctx)
}

func (c *TCPClient) connectWithReconnect(ctx context.Context) error {
	if c.stopping.Load() {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "client is stopping")
	}

	if err := c.doConnect(ctx); err != nil {
		log.Warnf("initial connection to %s failed: %v, will retry...", c.addr, err)
		c.startReconnect()
		return nil
	}

	c.wg.Add(1)
	routine.GoV2(func() error {
		defer c.wg.Done() // ✅ 确保 wg 减一
		c.healthCheck()
		return nil
	})

	return nil
}

func (c *TCPClient) doConnect(_ context.Context) error {
	conn, err := net.DialTimeout("tcp", c.addr, c.ConnectTimeout)
	if err != nil {
		return customerr.Wrap(err, int32(protocol.ErrorCode_Internal_Error))
	}

	c.mu.Lock()
	c.conn = NewTCPConnection(conn)
	c.mu.Unlock()

	c.connected.Store(true)
	c.reconnecting.Store(false)

	log.Infof("connected to %s", c.addr)

	// reset retry counter here for all successful connections
	c.retryCount.Store(0)

	c.startReceiveLoop()

	if c.onConnected != nil {
		routine.Run(func() {
			c.onConnected(c.conn)
		})
	}

	return nil
}

func (c *TCPClient) startReceiveLoop() {
	c.receiveMu.Lock()
	defer c.receiveMu.Unlock()

	if c.receiveCancel != nil {
		c.receiveCancel()
	}

	c.receiveCtx, c.receiveCancel = context.WithCancel(c.ctx)

	c.wg.Add(1)
	routine.GoV2(func() error {
		defer c.wg.Done() // ✅ 确保 wg 减一
		c.receiveLoop(c.receiveCtx)
		return nil
	})

	log.Debugf("receive loop started for %s", c.addr)
}

func (c *TCPClient) stopReceiveLoop() {
	c.receiveMu.Lock()
	defer c.receiveMu.Unlock()

	if c.receiveCancel != nil {
		c.receiveCancel()
		c.receiveCancel = nil
	}

	log.Debugf("receive loop stopped for %s", c.addr)
}

func (c *TCPClient) receiveLoop(ctx context.Context) {
	defer log.Debugf("receive loop exited for %s", c.addr)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		default:
		}

		if !c.connected.Load() {
			return
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		// ✅ 添加接收超时
		msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		msg, err := conn.ReceiveMessage(msgCtx)
		cancel()

		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				continue
			}
			log.Errorf("receive message failed: %v", err)
			c.handleDisconnect()
			return
		}

		// ✅ 添加处理超时
		handleCtx, handleCancel := context.WithTimeout(ctx, 10*time.Second)
		if err := c.handler.HandleMessage(handleCtx, conn, msg); err != nil {
			log.Errorf("handle message failed: %v", err)
		}
		handleCancel()
	}
}

func (c *TCPClient) healthCheck() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			if !c.connected.Load() {
				continue
			}

			if err := c.sendHeartbeat(); err != nil {
				log.Warnf("heartbeat to %s failed: %v, connection lost", c.addr, err)
				c.handleDisconnect()
			}
		}
	}
}

func (c *TCPClient) sendHeartbeat() error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	msg := GetMessage()
	defer PutMessage(msg)
	msg.Type = MsgTypeHeartbeat
	msg.Payload = []byte("ping")

	return conn.SendMessage(msg)
}

func (c *TCPClient) handleDisconnect() {
	if !c.connected.CompareAndSwap(true, false) {
		return
	}

	log.Warnf("connection to %s lost", c.addr)
	c.stopReceiveLoop()

	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	if c.onDisconnected != nil {
		routine.Run(
			func() {
				c.onDisconnected(c.conn)
			})
	}

	if c.EnableReconnect && !c.stopping.Load() {
		c.startReconnect()
	}
}

func (c *TCPClient) startReconnect() {
	if !c.reconnecting.CompareAndSwap(false, true) {
		return
	}

	c.wg.Add(1)
	routine.GoV2(func() error {
		defer c.wg.Done() // ✅ 确保 wg 减一
		c.reconnectLoop()
		return nil
	})
}

func (c *TCPClient) reconnectLoop() {
	defer c.reconnecting.Store(false)

	interval := c.InitialInterval

	for {
		if c.stopping.Load() {
			return
		}

		retries := c.retryCount.Add(1)
		if c.MaxRetries > 0 && int(retries) > c.MaxRetries {
			log.Errorf("max retries (%d) reached for %s, giving up", c.MaxRetries, c.addr)
			return
		}

		log.Infof("attempting to reconnect to %s (attempt %d)...", c.addr, retries)

		// 🔹每次重连使用独立短期拨号 context，防止阻塞
		dialCtx, cancel := context.WithTimeout(c.ctx, c.ConnectTimeout)
		err := c.doConnect(dialCtx)
		cancel()

		if err != nil {
			log.Warnf("reconnect to %s failed: %v, retry in %v", c.addr, err, interval)
			select {
			case <-c.ctx.Done():
				return
			case <-c.stopChan:
				return
			case <-time.After(interval):
				interval = time.Duration(float64(interval) * c.Multiplier)
				if interval > c.MaxInterval {
					interval = c.MaxInterval
				}
				continue
			}
		}

		// 🔹成功后清零重试次数
		c.retryCount.Store(0)

		log.Infof("reconnected to %s successfully after %d attempts", c.addr, retries)

		c.wg.Add(1)
		routine.GoV2(func() error {
			defer c.wg.Done() // ✅ 确保 wg 减一
			c.healthCheck()
			return nil
		})

		return
	}
}

func (c *TCPClient) SendMessage(msg *Message) error {
	if !c.connected.Load() {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Network_Timeout), "not connected to "+c.addr)
	}

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Network_Timeout), "connection is nil")
	}

	return conn.SendMessage(msg)
}

func (c *TCPClient) Close() error {
	if !c.stopping.CompareAndSwap(false, true) {
		return nil
	}

	log.Infof("closing TCP client to %s...", c.addr)
	c.stopReceiveLoop()

	if c.cancel != nil {
		c.cancel()
	}
	close(c.stopChan)

	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	c.connected.Store(false)

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Infof("TCP client to %s closed", c.addr)
	case <-time.After(10 * time.Second):
		log.Warnf("timeout waiting for TCP client to %s to close", c.addr)
	}

	return nil
}

func (c *TCPClient) GetConnection() IConnection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *TCPClient) IsConnected() bool {
	return c.connected.Load()
}

func (c *TCPClient) IsReconnecting() bool {
	return c.reconnecting.Load()
}
