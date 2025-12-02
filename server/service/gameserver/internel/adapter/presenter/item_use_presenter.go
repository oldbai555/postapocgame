package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
)

// ItemUsePresenter 物品使用响应构建器
type ItemUsePresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewItemUsePresenter 创建物品使用响应构建器
func NewItemUsePresenter(networkGateway gateway.NetworkGateway) *ItemUsePresenter {
	return &ItemUsePresenter{
		networkGateway: networkGateway,
	}
}

// SendUseItemResult 发送使用物品结果
func (p *ItemUsePresenter) SendUseItemResult(ctx context.Context, sessionID string, resp *protocol.S2CUseItemResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CUseItemResult), resp)
}
