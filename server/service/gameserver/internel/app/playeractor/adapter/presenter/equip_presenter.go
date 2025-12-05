package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/gateway"
)

// EquipPresenter 装备响应构建器
type EquipPresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewEquipPresenter 创建装备响应构建器
func NewEquipPresenter(networkGateway gateway.NetworkGateway) *EquipPresenter {
	return &EquipPresenter{
		networkGateway: networkGateway,
	}
}

// SendEquipResult 发送装备结果
func (p *EquipPresenter) SendEquipResult(ctx context.Context, sessionID string, resp *protocol.S2CEquipResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CEquipResult), resp)
}
