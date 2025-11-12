package gameserverlink

import (
	"context"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/base"
	"postapocgame/server/service/dungeonserver/internel/dshare"
)

// NetworkHandler 网络消息处理器（优化版）
type NetworkHandler struct {
	codec *network.Codec
}

// NewNetworkHandler 创建网络处理器
func NewNetworkHandler() *NetworkHandler {
	return &NetworkHandler{
		codec: network.DefaultCodec(),
	}
}

// HandleMessage 处理网络消息
func (h *NetworkHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	// 🔧 移到开头，确保握手后立即注册
	if msg.Type == network.MsgTypeHandshake {
		// 握手消息已经在 TCPServer 中处理
		return nil
	}

	// 注册GameServer连接
	h.registerGameServer(conn)

	switch msg.Type {
	case network.MsgTypeRPCRequest:
		return h.handleRPCRequest(ctx, conn, msg)
	case network.MsgTypeClient:
		return h.handleClientMsg(ctx, msg)
	case network.MsgTypeHeartbeat:
		return h.handleHeartbeat(conn, msg)
	default:
		log.Warnf("unknown message type: %d", msg.Type)
		return nil
	}
}

// registerGameServer 注册GameServer连接
func (h *NetworkHandler) registerGameServer(conn network.IConnection) {
	meta := conn.GetMeta()
	if meta == nil {
		return
	}

	handshake, ok := meta.(*network.HandshakeMessage)
	if !ok {
		return
	}

	// 注册到消息发送器
	GetMessageSender().RegisterGameServer(handshake.PlatformId, handshake.ZoneId, conn)

	log.Infof("GameServer registered: PlatformId=%d, ZoneId=%d", handshake.PlatformId, handshake.ZoneId)
}

// handleRPCRequest 处理来自GameServer的RPC请求
func (h *NetworkHandler) handleRPCRequest(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	req, err := h.codec.DecodeRPCRequest(msg.Payload)
	if err != nil {
		return customerr.Wrap(err)
	}

	log.Debugf("Received RPC Request: RequestId=%d, MsgId=%d", req.RequestId, req.MsgId)

	// 记录会话路由
	if req.SessionId != "" {
		meta := conn.GetMeta()
		if handshake, ok := meta.(*network.HandshakeMessage); ok {
			GetMessageSender().RegisterSessionRoute(req.SessionId, handshake.PlatformId, handshake.ZoneId)
		}
	}

	message := base.NewSessionMessage()
	message.SessionId = req.SessionId
	message.MsgId = req.MsgId
	message.Data = req.Data
	message.Context = ctx

	// 发送到Actor系统处理
	if err := dshare.SendMessageAsync(req.SessionId, message); err != nil {
		log.Errorf("send to actor failed: %v", err)
		return err
	}
	return nil
}

// handleClientMsg 处理客户端消息
func (h *NetworkHandler) handleClientMsg(ctx context.Context, msg *network.Message) error {
	fwdMsg, err := h.codec.DecodeForwardMessage(msg.Payload)
	if err != nil {
		return customerr.Wrap(err)
	}

	log.Debugf("Received Forward Message: SessionId=%s", fwdMsg.SessionId)

	message := base.NewSessionMessage()
	message.SessionId = fwdMsg.SessionId
	message.MsgId = dshare.DoNetWorkMsg
	message.Data = fwdMsg.Payload
	message.Context = ctx

	// 发送到Actor系统处理
	if err := dshare.SendMessageAsync(fwdMsg.SessionId, message); err != nil {
		log.Errorf("send to actor failed: %v", err)
		return err
	}
	return nil
}

// handleHeartbeat 处理心跳消息
func (h *NetworkHandler) handleHeartbeat(conn network.IConnection, msg *network.Message) error {
	pong := network.GetMessage()
	defer network.PutMessage(pong)

	pong.Type = network.MsgTypeHeartbeat
	pong.Payload = []byte("pong")

	return conn.SendMessage(pong)
}

// sendRPCResponse 发送RPC响应
func (h *NetworkHandler) sendRPCResponse(conn network.IConnection, resp *network.RPCResponse) error {
	rpcBuf := h.codec.EncodeRPCResponse(resp)
	defer network.PutBuffer(rpcBuf)

	msg := network.GetMessage()
	defer network.PutMessage(msg)

	msg.Type = network.MsgTypeRPCResponse
	msg.Payload = rpcBuf

	return conn.SendMessage(msg)
}

// CallGameServer 调用GameServer RPC
func (h *NetworkHandler) CallGameServer(ctx context.Context, sessionId string, msgId uint16, data []byte) error {
	conn, ok := GetMessageSender().GetGameServerBySession(sessionId)
	if !ok {
		return customerr.NewCustomErr("game server not found for session: %s", sessionId)
	}

	req := network.GetRPCRequest()
	defer network.PutRPCRequest(req)
	req.MsgId = msgId
	req.Data = data

	// 发送RPC请求
	if err := conn.SendRPCRequest(req); err != nil {
		return customerr.Wrap(err)
	}
	return nil
}

// Close 关闭处理器
func (h *NetworkHandler) Close() error {
	return nil
}
