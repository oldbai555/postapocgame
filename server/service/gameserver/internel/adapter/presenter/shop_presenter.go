package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
)

// ShopPresenter 商城响应构建器
type ShopPresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewShopPresenter 创建商城响应构建器
func NewShopPresenter(networkGateway gateway.NetworkGateway) *ShopPresenter {
	return &ShopPresenter{
		networkGateway: networkGateway,
	}
}

// SendShopBuyResult 发送购买结果
func (p *ShopPresenter) SendShopBuyResult(ctx context.Context, sessionID string, resp *protocol.S2CShopBuyResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CShopBuyResult), resp)
}
