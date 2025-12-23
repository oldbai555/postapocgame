package presenter

import (
	"context"
	"postapocgame/server/service/gameserver/internel/playeractor/gateway"
	"postapocgame/server/service/gameserver/internel/playeractor/service/playerauth"

	"postapocgame/server/internal/protocol"
)

type PlayerAuthPresenter struct {
	network gateway.NetworkGateway
}

func NewPlayerAuthPresenter(network gateway.NetworkGateway) *PlayerAuthPresenter {
	return &PlayerAuthPresenter{network: network}
}

func (p *PlayerAuthPresenter) S2CRegister(_ context.Context, sessionID string, result *playerauth.RegisterResult) error {
	resp := &protocol.S2CRegisterReq{
		Success: result.Success,
		Message: result.Message,
		Token:   result.Token,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CRegister), resp)
}

func (p *PlayerAuthPresenter) S2CLogin(_ context.Context, sessionID string, result *playerauth.LoginResult) error {
	resp := &protocol.S2CLoginReq{
		Success: result.Success,
		Message: result.Message,
		Token:   result.Token,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CLogin), resp)
}
