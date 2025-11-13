package gameserverlink

import (
	"context"
	"errors"
	"sync"

	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
)

var (
	ErrNotConnected = errors.New("not connected to game service")
)

// GameMessageHandler 消息处理器
type GameMessageHandler struct {
	codec    *network.Codec
	recvChan chan *network.ForwardMessage
	mu       sync.Mutex
	closed   bool
}

func NewGameMessageHandler() *GameMessageHandler {
	return &GameMessageHandler{
		codec:    network.DefaultCodec(),
		recvChan: make(chan *network.ForwardMessage, 1024),
	}
}

func (h *GameMessageHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	if msg.Type != network.MsgTypeClient {
		log.Debugf("ignore message type: %d", msg.Type)
		return nil
	}

	forwardMsg, err := h.codec.DecodeForwardMessage(msg.Payload)
	if err != nil {
		log.Errorf("DecodeForwardMessage failed: %v", err)
		return err
	}

	// 非阻塞发送到接收通道
	select {
	case h.recvChan <- forwardMsg:
		return nil
	default:
		log.Warnf("receive channel full, drop message for session: %s", forwardMsg.SessionId)
		return nil
	}
}

func (h *GameMessageHandler) ReceiveMessage(ctx context.Context) (*network.ForwardMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-h.recvChan:
		if !ok {
			return nil, errors.New("receive channel closed")
		}
		return msg, nil
	}
}

func (h *GameMessageHandler) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.closed {
		close(h.recvChan)
		h.closed = true
	}
}

// GameClient GameServer客户端
type GameClient struct {
	client  network.ITCPClient
	codec   *network.Codec
	handler *GameMessageHandler
	mu      sync.Mutex
	stopped bool
}

func NewGameClient(addr string) *GameClient {
	handler := NewGameMessageHandler()

	client := network.NewTCPClient(
		network.WithTCPClientOptionOnConn(func(conn network.IConnection) {
			log.Infof("connected to game server: %s", addr)
		}),
		network.WithTCPClientOptionOnDisConn(func(conn network.IConnection) {
			log.Warnf("disconnected from game server: %s", addr)
		}),
		network.WithTCPClientOptionNetworkMessageHandler(handler),
	)

	gsc := &GameClient{
		client:  client,
		codec:   network.DefaultCodec(),
		handler: handler,
	}

	return gsc
}

// Connect 连接到GameServer
func (gsc *GameClient) Connect(ctx context.Context, addr string) error {
	// 连接(会自动重连和启动接收协程)
	return gsc.client.Connect(ctx, addr)
}

// NotifySessionEvent 通知会话事件
func (gsc *GameClient) NotifySessionEvent(_ context.Context, event *network.SessionEvent) error {
	if !gsc.client.IsConnected() {
		return ErrNotConnected
	}

	payload := gsc.codec.EncodeSessionEvent(event)
	defer network.PutBuffer(payload)

	msg := network.GetMessage()
	defer network.PutMessage(msg)
	msg.Type = network.MsgTypeSessionEvent
	msg.Payload = payload

	return gsc.client.SendMessage(msg)
}

// ForwardClientMsg 转发客户端消息
func (gsc *GameClient) ForwardClientMsg(_ context.Context, forwardMsg *network.ForwardMessage) error {
	if !gsc.client.IsConnected() {
		return ErrNotConnected
	}

	data := gsc.codec.EncodeForwardMessage(forwardMsg)
	defer network.PutBuffer(data)

	msg := network.GetMessage()
	defer network.PutMessage(msg)
	msg.Type = network.MsgTypeClient
	msg.Payload = data

	return gsc.client.SendMessage(msg)
}

// ReceiveGsMessage 接收GameServer消息
func (gsc *GameClient) ReceiveGsMessage(ctx context.Context) (*network.ForwardMessage, error) {
	return gsc.handler.ReceiveMessage(ctx)
}

// Close 关闭连接
func (gsc *GameClient) Close() error {
	gsc.mu.Lock()
	if gsc.stopped {
		gsc.mu.Unlock()
		return nil
	}
	gsc.stopped = true
	gsc.mu.Unlock()

	log.Infof("closing game server connector...")

	// 关闭TCP客户端(会停止重连和接收协程)
	if err := gsc.client.Close(); err != nil {
		log.Errorf("close tcp client failed: %v", err)
	}

	// 关闭消息处理器
	gsc.handler.Close()

	log.Infof("game server connector closed")
	return nil
}
