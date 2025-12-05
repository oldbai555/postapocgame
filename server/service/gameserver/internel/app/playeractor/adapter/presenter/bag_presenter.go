package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/gateway"
)

// BagPresenter 背包响应构建器
type BagPresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewBagPresenter 创建背包响应构建器
func NewBagPresenter(networkGateway gateway.NetworkGateway) *BagPresenter {
	return &BagPresenter{
		networkGateway: networkGateway,
	}
}

// SendBagData 发送背包数据
func (p *BagPresenter) SendBagData(ctx context.Context, sessionID string, bagData *protocol.SiBagData) error {
	resp := &protocol.S2CBagDataReq{
		BagData: bagData,
	}
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CBagData), resp)
}
