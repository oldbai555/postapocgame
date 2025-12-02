package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
)

// AuctionPresenter 拍卖行响应
type AuctionPresenter struct {
	network gateway.NetworkGateway
}

func NewAuctionPresenter(network gateway.NetworkGateway) *AuctionPresenter {
	return &AuctionPresenter{network: network}
}

func (p *AuctionPresenter) SendError(ctx context.Context, sessionID string, message string) error {
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
		Code: -1,
		Msg:  message,
	})
}
