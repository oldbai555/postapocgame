package gateway

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/service/gameserver/internel/infrastructure/gatewaylink"
)

// NetworkGatewayImpl 网络网关实现
type NetworkGatewayImpl struct{}

// NewNetworkGateway 创建网络网关
func NewNetworkGateway() NetworkGateway {
	return &NetworkGatewayImpl{}
}

// SendToSession 发送消息到 Session
func (g *NetworkGatewayImpl) SendToSession(sessionID string, msgID uint16, data []byte) error {
	return gatewaylink.SendToSession(sessionID, msgID, data)
}

// SendToSessionProto 发送 Proto 消息到 Session
func (g *NetworkGatewayImpl) SendToSessionProto(sessionID string, msgID uint16, message proto.Message) error {
	return gatewaylink.SendToSessionProto(sessionID, msgID, message)
}
