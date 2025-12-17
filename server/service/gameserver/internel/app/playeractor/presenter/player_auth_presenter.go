package presenter

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/gateway"
	playerauth2 "postapocgame/server/service/gameserver/internel/app/playeractor/service/playerauth"

	"postapocgame/server/internal/protocol"
)

// PlayerAuthPresenter 账号注册/登录回包
type PlayerAuthPresenter struct {
	network gateway.NetworkGateway
}

// NewPlayerAuthPresenter 创建 Presenter
func NewPlayerAuthPresenter(network gateway.NetworkGateway) *PlayerAuthPresenter {
	return &PlayerAuthPresenter{network: network}
}

// SendRegisterResult 回包注册结果
func (p *PlayerAuthPresenter) SendRegisterResult(ctx context.Context, sessionID string, result *playerauth2.RegisterResult) error {
	resp := &protocol.S2CRegisterResultReq{
		Success: result.Success,
		Message: result.Message,
		Token:   result.Token,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CRegisterResult), resp)
}

// SendLoginResult 回包登录结果
func (p *PlayerAuthPresenter) SendLoginResult(ctx context.Context, sessionID string, result *playerauth2.LoginResult) error {
	resp := &protocol.S2CLoginResultReq{
		Success: result.Success,
		Message: result.Message,
		Token:   result.Token,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CLoginResult), resp)
}
