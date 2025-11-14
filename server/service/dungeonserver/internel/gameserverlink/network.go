package gameserverlink

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/dshare"
)

// NetworkHandler 网络消息处理器（优化版）
type NetworkHandler struct{}

// NewNetworkHandler 创建网络处理器
func NewNetworkHandler() *NetworkHandler {
	return &NetworkHandler{}
}

// HandleMessage 处理网络消息
func (h *NetworkHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	switch msg.Type {
	case network.MsgTypeHandshake:
		return h.handleHandshake(conn, msg)
	case network.MsgTypeRPCRequest:
		return h.handleRPCRequest(ctx, msg)
	case network.MsgTypeClient:
		return h.handleClientMsg(ctx, msg)
	case network.MsgTypeHeartbeat:
		return h.handleHeartbeat(conn)
	default:
		log.Warnf("unknown message type: %d", msg.Type)
		return nil
	}
}

// handleRPCRequest 处理来自GameServer的RPC请求
func (h *NetworkHandler) handleRPCRequest(ctx context.Context, msg *network.Message) error {
	req, err := dshare.Codec.DecodeRPCRequest(msg.Payload)
	if err != nil {
		return customerr.Wrap(err)
	}

	log.Debugf("Received RPC Request: RequestId=%d, MsgId=%d", req.RequestId, req.MsgId)

	newCtx := context.WithValue(ctx, dshare.ContextKeySession, req.SessionId)
	message := actor.NewBaseMessage(newCtx, dshare.DoRpcMsg, msg.Payload)

	// 发送到Actor系统处理
	if err := dshare.SendMessageAsync(req.SessionId, message); err != nil {
		log.Errorf("send to actor failed: %v", err)
		return err
	}
	return nil
}

// handleClientMsg 处理客户端消息
func (h *NetworkHandler) handleClientMsg(ctx context.Context, msg *network.Message) error {
	fwdMsg, err := dshare.Codec.DecodeForwardMessage(msg.Payload)
	if err != nil {
		return customerr.Wrap(err)
	}

	log.Debugf("Received Forward Message: SessionId=%s", fwdMsg.SessionId)

	newCtx := context.WithValue(ctx, dshare.ContextKeySession, fwdMsg.SessionId)
	message := actor.NewBaseMessage(newCtx, dshare.DoNetWorkMsg, fwdMsg.Payload)

	// 发送到Actor系统处理
	if err := dshare.SendMessageAsync(fwdMsg.SessionId, message); err != nil {
		log.Errorf("send to actor failed: %v", err)
		return err
	}
	return nil
}

// handleHeartbeat 处理心跳消息
func (h *NetworkHandler) handleHeartbeat(conn network.IConnection) error {
	pong := network.GetMessage()
	defer network.PutMessage(pong)

	pong.Type = network.MsgTypeHeartbeat
	pong.Payload = []byte("pong")

	return conn.SendMessage(pong)
}

// sendRPCResponse 发送RPC响应
func (h *NetworkHandler) sendRPCResponse(conn network.IConnection, resp *network.RPCResponse) error {
	rpcBuf := dshare.Codec.EncodeRPCResponse(resp)
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
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "game server not found for session: %s", sessionId)
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

func (h *NetworkHandler) handleHandshake(conn network.IConnection, msg *network.Message) error {
	var req protocol.G2DSyncGameDataReq
	if err := internal.Unmarshal(msg.Payload, &req); err != nil {
		return customerr.Wrap(err)
	}
	GetMessageSender().RegisterGameServer(req.PlatformId, req.SrvId, conn)
	log.Infof("game server connected: %d %d", req.PlatformId, req.SrvId)
	return nil
}
