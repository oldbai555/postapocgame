package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/gateway"
)

// QuestPresenter 任务响应构建器
type QuestPresenter struct {
	networkGateway gateway.NetworkGateway
}

// NewQuestPresenter 创建任务响应构建器
func NewQuestPresenter(networkGateway gateway.NetworkGateway) *QuestPresenter {
	return &QuestPresenter{
		networkGateway: networkGateway,
	}
}

// SendTalkToNPCResult 发送NPC对话结果
func (p *QuestPresenter) SendTalkToNPCResult(ctx context.Context, sessionID string, resp *protocol.S2CTalkToNPCResultReq) error {
	return p.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CTalkToNPCResult), resp)
}
