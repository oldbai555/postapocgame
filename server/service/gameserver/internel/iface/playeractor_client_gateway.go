package iface

import "google.golang.org/protobuf/proto"

// NetworkGateway 网络网关接口（玩家侧向客户端发包）
type NetworkGateway interface {
	// SendToSession 发送消息到 Session
	SendToSession(sessionID string, msgID uint16, data []byte) error

	// SendToSessionProto 发送 Proto 消息到 Session
	SendToSessionProto(sessionID string, msgID uint16, message proto.Message) error
}

// SessionGateway 会话访问接口
type SessionGateway interface {
	GetSession(sessionID string) ISession
}

// ClientGateway 统一的客户端通信与会话访问网关（合并 NetworkGateway + SessionGateway）
type ClientGateway interface {
	NetworkGateway
	SessionGateway
}
