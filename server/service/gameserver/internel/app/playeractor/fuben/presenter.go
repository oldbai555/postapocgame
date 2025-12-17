package fuben

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/iface"
)

// FubenPresenter 副本响应构建器
type FubenPresenter struct {
	networkGateway iface.NetworkGateway
}

// NewFubenPresenter 创建副本响应构建器
func NewFubenPresenter(networkGateway iface.NetworkGateway) *FubenPresenter {
	return &FubenPresenter{
		networkGateway: networkGateway,
	}
}

// SendEnterDungeonResult 发送进入副本结果
func (p *FubenPresenter) SendEnterDungeonResult(ctx context.Context, sessionID string, resp *protocol.S2CEnterDungeonResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
}
