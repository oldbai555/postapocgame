package gameserverlink

import (
	"context"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"sync"
	"time"
)

// MessageSender DungeonServer消息发送器
type MessageSender struct {
	gameServers   map[argsdef.GameServerKey]network.IMessageSender
	sessionRoutes map[string]argsdef.GameServerKey
	codec         *network.Codec
	mu            sync.RWMutex

	// 同步RPC调用等待响应
	pendingRPCs   map[uint32]chan *network.RPCResponse
	rpcMu         sync.RWMutex
	nextRequestId uint32

	// 异步RPC的SessionId映射（RequestId -> SessionId）
	asyncRpcSessions map[uint32]string
	asyncRpcMu       sync.RWMutex
}

var (
	globalSender *MessageSender
	senderOnce   sync.Once
)

// GetMessageSender 获取全局消息发送器
func GetMessageSender() *MessageSender {
	senderOnce.Do(func() {
		globalSender = &MessageSender{
			gameServers:      make(map[argsdef.GameServerKey]network.IMessageSender),
			sessionRoutes:    make(map[string]argsdef.GameServerKey),
			codec:            network.DefaultCodec(),
			pendingRPCs:      make(map[uint32]chan *network.RPCResponse),
			nextRequestId:    1,
			asyncRpcSessions: make(map[uint32]string),
		}
	})
	return globalSender
}

// RegisterGameServer 注册GameServer连接
func (s *MessageSender) RegisterGameServer(platformId, srvId uint32, conn network.IConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := argsdef.GameServerKey{
		PlatformId: platformId,
		SrvId:      srvId,
	}
	sender := network.NewBaseMessageSender(conn)
	s.gameServers[key] = sender
}

// RegisterSessionRoute 注册会话路由
func (s *MessageSender) RegisterSessionRoute(sessionId string, platformId, srvId uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := argsdef.GameServerKey{
		PlatformId: platformId,
		SrvId:      srvId,
	}
	s.sessionRoutes[sessionId] = key
}

// GetGameServerBySession 通过SessionId获取GameServer连接
func (s *MessageSender) GetGameServerBySession(sessionId string) (network.IMessageSender, bool) {
	s.mu.RLock()
	key, ok := s.sessionRoutes[sessionId]
	s.mu.RUnlock()

	if !ok {
		return nil, false
	}

	s.mu.RLock()
	conn, ok := s.gameServers[key]
	s.mu.RUnlock()

	return conn, ok
}

// SendToClient 发送消息给客户端
func (s *MessageSender) SendToClient(sessionId string, msgId uint16, data []byte) error {
	// 获取GameServer连接
	conn, ok := s.GetGameServerBySession(sessionId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no gameserver connection for session: %s", sessionId)
	}
	return conn.SendToClient(sessionId, msgId, data)
}

// SendToClientJSON 发送JSON消息给客户端
func (s *MessageSender) SendToClientJSON(sessionId string, msgId uint16, v interface{}) error {
	// 获取GameServer连接
	conn, ok := s.GetGameServerBySession(sessionId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no gameserver connection for session: %s", sessionId)
	}
	return conn.SendToClientJSON(sessionId, msgId, v)
}

// RemoveSession 移除会话路由
func (s *MessageSender) RemoveSession(sessionId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessionRoutes, sessionId)
}

// SendToClient 全局函数：发送消息给客户端
func SendToClient(sessionId string, msgId uint16, data []byte) error {
	return GetMessageSender().SendToClient(sessionId, msgId, data)
}

// SendToClientJSON 全局函数：发送JSON消息给客户端
func SendToClientJSON(sessionId string, msgId uint16, v interface{}) error {
	return GetMessageSender().SendToClientJSON(sessionId, msgId, v)
}

// GetFirstGameServer 获取第一个可用的GameServer连接
func (s *MessageSender) GetFirstGameServer() (network.IMessageSender, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, conn := range s.gameServers {
		return conn, true
	}
	return nil, false
}

// CallGameServer 调用GameServer RPC（异步）
func (s *MessageSender) CallGameServer(ctx context.Context, sessionId string, msgId uint16, data []byte) error {
	var conn network.IMessageSender
	var ok bool

	if sessionId != "" {
		// 如果指定了sessionId,使用session路由
		conn, ok = s.GetGameServerBySession(sessionId)
	} else {
		// 否则使用第一个可用的GameServer
		conn, ok = s.GetFirstGameServer()
	}

	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no gameserver connection available")
	}

	// 生成RequestId
	s.rpcMu.Lock()
	requestId := s.nextRequestId
	s.nextRequestId++
	s.rpcMu.Unlock()

	// 如果是异步RPC（如D2GAddItem），保存SessionId映射
	if msgId == uint16(protocol.D2GRpcProtocol_D2GAddItem) && sessionId != "" {
		s.asyncRpcMu.Lock()
		s.asyncRpcSessions[requestId] = sessionId
		s.asyncRpcMu.Unlock()
	}

	return conn.SendRPCRequest(&network.RPCRequest{
		RequestId: requestId,
		SessionId: sessionId,
		MsgId:     msgId,
		Data:      data,
	})
}

// CallGameServer 全局函数:调用GameServer RPC
func CallGameServer(ctx context.Context, sessionId string, msgId uint16, data []byte) error {
	return GetMessageSender().CallGameServer(ctx, sessionId, msgId, data)
}

// SyncCall 同步调用GameServer RPC并返回响应
func SyncCall(ctx context.Context, sessionId string, msgId uint16, data []byte) ([]byte, error) {
	return GetMessageSender().SyncCall(ctx, sessionId, msgId, data)
}

// SyncCall 同步调用GameServer RPC并返回响应
func (s *MessageSender) SyncCall(ctx context.Context, sessionId string, msgId uint16, data []byte) ([]byte, error) {
	var conn network.IMessageSender
	var ok bool

	if sessionId != "" {
		// 如果指定了sessionId,使用session路由
		conn, ok = s.GetGameServerBySession(sessionId)
	} else {
		// 否则使用第一个可用的GameServer
		conn, ok = s.GetFirstGameServer()
	}

	if !ok {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no gameserver connection available")
	}

	// 生成请求ID
	s.rpcMu.Lock()
	requestId := s.nextRequestId
	s.nextRequestId++
	// 创建响应channel
	respChan := make(chan *network.RPCResponse, 1)
	s.pendingRPCs[requestId] = respChan
	s.rpcMu.Unlock()

	// 发送RPC请求
	req := &network.RPCRequest{
		RequestId: requestId,
		SessionId: sessionId,
		MsgId:     msgId,
		Data:      data,
	}
	if err := conn.SendRPCRequest(req); err != nil {
		// 清理pending
		s.rpcMu.Lock()
		delete(s.pendingRPCs, requestId)
		s.rpcMu.Unlock()
		return nil, customerr.Wrap(err)
	}

	// 等待响应（超时5秒）
	select {
	case resp := <-respChan:
		// 清理pending
		s.rpcMu.Lock()
		delete(s.pendingRPCs, requestId)
		s.rpcMu.Unlock()

		if !resp.Success {
			return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), string(resp.Data))
		}
		return resp.Data, nil
	case <-time.After(5 * time.Second):
		// 超时，清理pending
		s.rpcMu.Lock()
		delete(s.pendingRPCs, requestId)
		s.rpcMu.Unlock()
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "RPC call timeout")
	case <-ctx.Done():
		// 上下文取消，清理pending
		s.rpcMu.Lock()
		delete(s.pendingRPCs, requestId)
		s.rpcMu.Unlock()
		return nil, ctx.Err()
	}
}

// HandleRPCResponse 处理RPC响应（由NetworkHandler调用）
func (s *MessageSender) HandleRPCResponse(resp *network.RPCResponse) {
	s.rpcMu.RLock()
	respChan, ok := s.pendingRPCs[resp.RequestId]
	s.rpcMu.RUnlock()

	if ok {
		select {
		case respChan <- resp:
		default:
			// channel已满，忽略
		}
	}
}
