package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
)

// GuildPresenter 公会系统响应构建
type GuildPresenter struct {
	network gateway.NetworkGateway
}

// NewGuildPresenter 创建 Presenter
func NewGuildPresenter(network gateway.NetworkGateway) *GuildPresenter {
	return &GuildPresenter{network: network}
}

func (p *GuildPresenter) SendCreateResult(ctx context.Context, sessionID string, success bool, message string) error {
	resp := &protocol.S2CCreateGuildResultReq{
		Success: success,
		Message: message,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CCreateGuildResult), resp)
}

func (p *GuildPresenter) SendJoinResult(ctx context.Context, sessionID string, success bool, message string) error {
	resp := &protocol.S2CJoinGuildResultReq{
		Success: success,
		Message: message,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CJoinGuildResult), resp)
}

func (p *GuildPresenter) SendLeaveResult(ctx context.Context, sessionID string, success bool, message string) error {
	resp := &protocol.S2CLeaveGuildResultReq{
		Success: success,
		Message: message,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), resp)
}

func (p *GuildPresenter) SendGuildInfo(ctx context.Context, sessionID string, guildData *protocol.GuildData) error {
	resp := &protocol.S2CGuildInfoReq{
		GuildData: guildData,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CGuildInfo), resp)
}

func (p *GuildPresenter) SendError(ctx context.Context, sessionID string, message string) error {
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
		Code: -1,
		Msg:  message,
	})
}
