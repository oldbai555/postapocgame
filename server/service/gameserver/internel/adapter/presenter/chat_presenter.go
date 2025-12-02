package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
)

// ChatPresenter 聊天系统响应构建器
type ChatPresenter struct {
	network gateway.NetworkGateway
}

func NewChatPresenter(network gateway.NetworkGateway) *ChatPresenter {
	return &ChatPresenter{network: network}
}

func (p *ChatPresenter) SendError(ctx context.Context, sessionID string, message string) error {
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
		Code: -1,
		Msg:  message,
	})
}
