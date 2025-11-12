package dungeonserverlink

import (
	"context"
	"fmt"
	"sync"
	"time"

	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/config"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
)

type RPCHandler func(ctx context.Context, sessionId string, data []byte) error

// DungeonMessageHandler 消息处理器
type DungeonMessageHandler struct {
	codec         *network.Codec
	rpcHandlers   map[uint16]RPCHandler
	rpcHandlersMu sync.RWMutex
}

func NewDungeonMessageHandler() *DungeonMessageHandler {
	return &DungeonMessageHandler{
		codec:       network.DefaultCodec(),
		rpcHandlers: make(map[uint16]RPCHandler),
	}
}

func (h *DungeonMessageHandler) RegisterRPCHandler(msgId uint16, handler RPCHandler) {
	h.rpcHandlersMu.Lock()
	defer h.rpcHandlersMu.Unlock()
	h.rpcHandlers[msgId] = handler
	log.Infof("RPC handler registered: msgId=%d", msgId)
}

func (h *DungeonMessageHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	switch msg.Type {
	case network.MsgTypeRPCRequest:
		return h.handleRPCRequest(ctx, msg)
	case network.MsgTypeHeartbeat:
		return nil
	case network.MsgTypeClient:
		return h.handleClientMessage(msg)
	default:
		log.Debugf("unknown message type from dungeon service: %d", msg.Type)
		return nil
	}
}

func (h *DungeonMessageHandler) handleRPCRequest(ctx context.Context, msg *network.Message) error {
	req, err := h.codec.DecodeRPCRequest(msg.Payload)
	if err != nil {
		log.Errorf("decode rpc request failed: %v", err)
		return err
	}

	log.Debugf("Received RPC Request from DungeonServer: RequestId=%d, MsgId=%d", req.RequestId, req.MsgId)

	h.rpcHandlersMu.RLock()
	handler, ok := h.rpcHandlers[req.MsgId]
	h.rpcHandlersMu.RUnlock()

	if !ok {
		log.Errorf("no rpc handler registered for msgId: %d", req.MsgId)
		return fmt.Errorf("no handler for msgId: %d", req.MsgId)
	}

	if err := handler(ctx, req.SessionId, req.Data); err != nil {
		log.Errorf("handle rpc request failed: %v", err)
		return err
	}

	return nil
}

func (h *DungeonMessageHandler) handleClientMessage(msg *network.Message) error {
	forwardMessage, err := h.codec.DecodeForwardMessage(msg.Payload)
	if err != nil {
		log.Errorf("decode forward message failed: %v", err)
		return err
	}

	if err := gatewaylink.ForwardClientMsg(forwardMessage.SessionId, forwardMessage.Payload); err != nil {
		log.Errorf("forward client message failed: %v", err)
		return err
	}

	return nil
}

// DungeonClient DungeonServer客户端
type DungeonClient struct {
	config    *config.ServerConfig
	connPools map[uint8]network.ITCPClient
	mu        sync.RWMutex

	codec   *network.Codec
	handler *DungeonMessageHandler
}

func NewDungeonClient(config *config.ServerConfig) *DungeonClient {
	return &DungeonClient{
		config:    config,
		connPools: make(map[uint8]network.ITCPClient),
		codec:     network.DefaultCodec(),
		handler:   NewDungeonMessageHandler(),
	}
}

func (dc *DungeonClient) RegisterRPCHandler(msgId uint16, handler RPCHandler) {
	dc.handler.RegisterRPCHandler(msgId, handler)
}

// Connect 连接到DungeonServer(使用自动重连)
func (dc *DungeonClient) Connect(ctx context.Context, srvType uint8, addr string) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if _, ok := dc.connPools[srvType]; ok {
		return fmt.Errorf("already connected to dungeon service: srvType=%d", srvType)
	}

	clientConfig := &network.TCPClientConfig{
		HandshakeEnable: true,
		Handshake: &network.HandshakeMessage{
			ServerType: 1, // 1=GameServer
			PlatformId: dc.config.PlatformID,
			ZoneId:     dc.config.SrvId,
			SrvType:    0,
		},
		EnableReconnect: true,
		ReconnectConfig: network.DefaultReconnectConfig(),
	}

	client := network.NewTCPClient(clientConfig, dc.handler)

	// 设置连接/断开回调
	client.SetCallbacks(
		func() {
			log.Infof("connected to DungeonServer: srvType=%d, addr=%s", srvType, addr)
		},
		func() {
			log.Warnf("disconnected from DungeonServer: srvType=%d, addr=%s", srvType, addr)
		},
	)

	dc.connPools[srvType] = client

	// 连接(会自动重连和启动接收协程)
	if err := client.Connect(ctx, addr); err != nil {
		return err
	}

	log.Infof("DungeonClient connected: srvType=%d, addr=%s", srvType, addr)
	return nil
}

// AsyncCall 异步调用DungeonServer
func (dc *DungeonClient) AsyncCall(ctx context.Context, srvType uint8, sessionId string, msgId uint16, data []byte) error {
	dc.mu.RLock()
	client, ok := dc.connPools[srvType]
	dc.mu.RUnlock()

	if !ok {
		return fmt.Errorf("dungeon service not connected: srvType=%d", srvType)
	}

	// 🔧 添加重试逻辑
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if !client.IsConnected() {
			if i < maxRetries-1 {
				time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
				continue
			}
			return fmt.Errorf("dungeon service not connected after %d retries: srvType=%d", maxRetries, srvType)
		}

		req := network.GetRPCRequest()
		req.SessionId = sessionId
		req.MsgId = msgId
		req.Data = data

		rpcBuf := dc.codec.EncodeRPCRequest(req)
		msg := network.GetMessage()
		msg.Type = network.MsgTypeRPCRequest
		msg.Payload = rpcBuf

		err := client.SendMessage(msg)

		network.PutMessage(msg)
		network.PutBuffer(rpcBuf)
		network.PutRPCRequest(req)

		if err == nil {
			log.Debugf("Async RPC: MsgId=%d, SessionId=%s", msgId, sessionId)
			return nil
		}

		if i < maxRetries-1 {
			log.Warnf("Send RPC failed (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
		}
	}

	return fmt.Errorf("failed to send RPC after %d retries", maxRetries)
}

// Close 关闭连接
func (dc *DungeonClient) Close() error {
	log.Infof("DungeonClient shutting down...")

	// 停止所有TCP客户端(会停止重连和接收协程)
	dc.mu.Lock()
	for srvType, client := range dc.connPools {
		if err := client.Close(); err != nil {
			log.Errorf("stop client failed: srvType=%d, err=%v", srvType, err)
		}
	}
	dc.mu.Unlock()

	log.Infof("DungeonClient shutdown complete")
	return nil
}
