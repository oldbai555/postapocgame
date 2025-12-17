package money

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/gateway"
)

// MoneyPresenter 货币响应构建器
type MoneyPresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewMoneyPresenter 创建货币响应构建器
func NewMoneyPresenter(networkGateway gateway.NetworkGateway) *MoneyPresenter {
	return &MoneyPresenter{
		networkGateway: networkGateway,
	}
}

// SendMoneyData 发送货币数据
func (p *MoneyPresenter) SendMoneyData(ctx context.Context, sessionID string, moneyData *protocol.SiMoneyData) error {
	resp := &protocol.S2CMoneyDataReq{
		MoneyData: moneyData,
	}
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CMoneyData), resp)
}
