package gameserverlink

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// MessageSender DungeonServer消息发送器
type MessageSender struct {
	gameServers   map[argsdef.GameServerKey]network.IMessageSender
	sessionRoutes map[string]argsdef.GameServerKey
	codec         *network.Codec
}

var (
	globalSender *MessageSender
)

// GetMessageSender 获取全局消息发送器
func GetMessageSender() *MessageSender {
	if globalSender == nil {
		globalSender = &MessageSender{
			gameServers:   make(map[argsdef.GameServerKey]network.IMessageSender),
			sessionRoutes: make(map[string]argsdef.GameServerKey),
			codec:         network.DefaultCodec(),
		}
	}
	return globalSender
}

// RegisterGameServer 注册GameServer连接
func (s *MessageSender) RegisterGameServer(platformId, srvId uint32, conn network.IConnection) {
	key := argsdef.GameServerKey{
		PlatformId: platformId,
		SrvId:      srvId,
	}
	sender := network.NewBaseMessageSender(conn)
	s.gameServers[key] = sender
}

// RegisterSessionRoute 注册会话路由
func (s *MessageSender) RegisterSessionRoute(sessionId string, platformId, srvId uint32) {
	key := argsdef.GameServerKey{
		PlatformId: platformId,
		SrvId:      srvId,
	}
	s.sessionRoutes[sessionId] = key
}

// GetGameServerBySession 通过SessionId获取GameServer连接
func (s *MessageSender) GetGameServerBySession(sessionId string) (network.IMessageSender, bool) {
	key, ok := s.sessionRoutes[sessionId]
	if !ok {
		return nil, false
	}
	conn, ok := s.gameServers[key]
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

// SendToClientProto 发送Proto消息给客户端
func (s *MessageSender) SendToClientProto(sessionId string, msgId uint16, message proto.Message) error {
	conn, ok := s.GetGameServerBySession(sessionId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no gameserver connection for session: %s", sessionId)
	}
	return conn.SendToClientProto(sessionId, msgId, message)
}

// RemoveSession 移除会话路由
func (s *MessageSender) RemoveSession(sessionId string) {
	delete(s.sessionRoutes, sessionId)
}

// SendToClient 全局函数：发送消息给客户端
func SendToClient(sessionId string, msgId uint16, data []byte) error {
	return GetMessageSender().SendToClient(sessionId, msgId, data)
}

// SendToClientProto 全局函数：发送Proto消息给客户端
func SendToClientProto(sessionId string, msgId uint16, message proto.Message) error {
	return GetMessageSender().SendToClientProto(sessionId, msgId, message)
}

// GetFirstGameServer 获取第一个可用的GameServer连接
func (s *MessageSender) GetFirstGameServer() (network.IMessageSender, bool) {
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

	return conn.SendRPCRequest(&network.RPCRequest{
		SessionId: sessionId,
		MsgId:     msgId,
		Data:      data,
	})
}

// CallGameServer 全局函数:调用GameServer RPC
func CallGameServer(ctx context.Context, sessionId string, msgId uint16, data []byte) error {
	return GetMessageSender().CallGameServer(ctx, sessionId, msgId, data)
}
