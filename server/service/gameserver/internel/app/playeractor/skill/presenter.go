package skill

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/gateway"
)

// SkillPresenter 技能响应构建器
type SkillPresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewSkillPresenter 创建技能响应构建器
func NewSkillPresenter(networkGateway gateway.NetworkGateway) *SkillPresenter {
	return &SkillPresenter{
		networkGateway: networkGateway,
	}
}

// SendLearnSkillResult 发送学习技能结果
func (p *SkillPresenter) SendLearnSkillResult(ctx context.Context, sessionID string, resp *protocol.S2CLearnSkillResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CLearnSkillResult), resp)
}

// SendUpgradeSkillResult 发送升级技能结果
func (p *SkillPresenter) SendUpgradeSkillResult(ctx context.Context, sessionID string, resp *protocol.S2CUpgradeSkillResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CUpgradeSkillResult), resp)
}
