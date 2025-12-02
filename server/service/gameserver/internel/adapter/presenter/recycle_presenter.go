package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
)

// RecyclePresenter 回收系统响应构建器
type RecyclePresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewRecyclePresenter 创建回收系统响应构建器
func NewRecyclePresenter(networkGateway gateway.NetworkGateway) *RecyclePresenter {
	return &RecyclePresenter{
		networkGateway: networkGateway,
	}
}

// SendRecycleResult 发送回收结果
func (p *RecyclePresenter) SendRecycleResult(ctx context.Context, sessionID string, resp *protocol.S2CRecycleItemResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CRecycleItemResult), resp)
}
