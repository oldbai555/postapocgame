package gateway

import (
	"google.golang.org/protobuf/proto"
	gatewaylink2 "postapocgame/server/service/gameserver/internel/gatewaylink"
)

// ClientGatewayImpl 合并网络发送与会话访问能力
type ClientGatewayImpl struct{}

func NewClientGateway() ClientGateway {
	return &ClientGatewayImpl{}
}

// SendToSession 发送消息到 Session
func (g *ClientGatewayImpl) SendToSession(sessionID string, msgID uint16, data []byte) error {
	return gatewaylink2.SendToSession(sessionID, msgID, data)
}

// SendToSessionProto 发送 Proto 消息到 Session
func (g *ClientGatewayImpl) SendToSessionProto(sessionID string, msgID uint16, message proto.Message) error {
	return gatewaylink2.SendToSessionProto(sessionID, msgID, message)
}

// GetSession 根据 SessionID 获取会话
func (g *ClientGatewayImpl) GetSession(sessionID string) Session {
	if sessionID == "" {
		return nil
	}
	return gatewaylink2.GetSession(sessionID)
}
