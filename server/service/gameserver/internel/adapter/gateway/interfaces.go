package gateway

import (
	"google.golang.org/protobuf/proto"
)

// NetworkGateway 网络网关接口（Adapter 层定义）
type NetworkGateway interface {
	// SendToSession 发送消息到 Session
	SendToSession(sessionID string, msgID uint16, data []byte) error

	// SendToSessionProto 发送 Proto 消息到 Session
	SendToSessionProto(sessionID string, msgID uint16, message proto.Message) error
}
