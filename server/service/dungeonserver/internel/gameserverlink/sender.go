package gameserverlink

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/customerr"
	"sync"
)

// MessageSender DungeonServer消息发送器
type MessageSender struct {
	gameServers   map[argsdef.GameServerKey]network.IMessageSender
	sessionRoutes map[string]argsdef.GameServerKey
	codec         *network.Codec
	mu            sync.RWMutex
}

var (
	globalSender *MessageSender
	senderOnce   sync.Once
)

// GetMessageSender 获取全局消息发送器
func GetMessageSender() *MessageSender {
	senderOnce.Do(func() {
		globalSender = &MessageSender{
			gameServers:   make(map[argsdef.GameServerKey]network.IMessageSender),
			sessionRoutes: make(map[string]argsdef.GameServerKey),
			codec:         network.DefaultCodec(),
		}
	})
	return globalSender
}

// RegisterGameServer 注册GameServer连接
func (s *MessageSender) RegisterGameServer(platformId, zoneId uint32, conn network.IConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := argsdef.GameServerKey{
		PlatformId: platformId,
		ZoneId:     zoneId,
	}
	sender := network.NewBaseMessageSender(conn)
	s.gameServers[key] = sender
}

// RegisterSessionRoute 注册会话路由
func (s *MessageSender) RegisterSessionRoute(sessionId string, platformId, zoneId uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := argsdef.GameServerKey{
		PlatformId: platformId,
		ZoneId:     zoneId,
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
		return customerr.NewCustomErr("no gameserver connection for session: %s", sessionId)
	}
	return conn.SendToClient(sessionId, msgId, data)
}

// SendToClientJSON 发送JSON消息给客户端
func (s *MessageSender) SendToClientJSON(sessionId string, msgId uint16, v interface{}) error {
	// 获取GameServer连接
	conn, ok := s.GetGameServerBySession(sessionId)
	if !ok {
		return customerr.NewCustomErr("no gameserver connection for session: %s", sessionId)
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
