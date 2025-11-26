package gameserverlink

import (
	"context"
	"postapocgame/server/pkg/customerr"
	"sync"

	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
)

var (
	ErrNotConnected = customerr.NewErrorByCode(-1, "not connected to game service")
)

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
