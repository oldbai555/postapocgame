package controller

import (
	"context"
	"postapocgame/server/service/gameserver/internel/adapter/system"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/quest"
)

// QuestController 任务控制器
type QuestController struct {
	acceptQuestUseCase    *quest.AcceptQuestUseCase
	updateProgressUseCase *quest.UpdateQuestProgressUseCase
	submitQuestUseCase    *quest.SubmitQuestUseCase
	presenter             *presenter.QuestPresenter
}

// NewQuestController 创建任务控制器
func NewQuestController() *QuestController {
	container := di.GetContainer()
	acceptQuestUC := quest.NewAcceptQuestUseCase(container.PlayerGateway(), container.ConfigGateway())
	updateProgressUC := quest.NewUpdateQuestProgressUseCase(container.PlayerGateway(), container.ConfigGateway())
	submitQuestUC := quest.NewSubmitQuestUseCase(container.PlayerGateway(), container.ConfigGateway())

	// 注入依赖
	levelUseCase := system.NewLevelUseCaseAdapter()
	rewardUseCase := usecaseadapter.NewRewardUseCaseAdapter()
	dailyActivityUseCase := system.NewDailyActivityUseCaseAdapter()
	acceptQuestUC.SetDependencies(levelUseCase)
	submitQuestUC.SetDependencies(levelUseCase, rewardUseCase, dailyActivityUseCase)

	return &QuestController{
		acceptQuestUseCase:    acceptQuestUC,
		updateProgressUseCase: updateProgressUC,
		submitQuestUseCase:    submitQuestUC,
		presenter:             presenter.NewQuestPresenter(container.NetworkGateway()),
	}
}

// HandleTalkToNPC 处理NPC对话请求
func (c *QuestController) HandleTalkToNPC(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2STalkToNPCReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 检查NPC配置
	npcConfigRaw := di.GetContainer().ConfigGateway().GetNPCSceneConfig(req.NpcId)
	if npcConfigRaw == nil {
		resp := &protocol.S2CTalkToNPCResultReq{
			Success: false,
			Message: "NPC不存在",
			NpcId:   req.NpcId,
		}
		return c.presenter.SendTalkToNPCResult(ctx, sessionID, resp)
	}

	// 触发任务事件（和NPC对话）
	if err := c.updateProgressUseCase.Execute(ctx, roleID, uint32(protocol.QuestType_QuestTypeTalkToNPC), req.NpcId, 1); err != nil {
		log.Warnf("Update quest progress failed: %v", err)
		// 不返回错误，继续处理对话
	}

	// 发送对话结果
	resp := &protocol.S2CTalkToNPCResultReq{
		Success: true,
		Message: "对话成功",
		NpcId:   req.NpcId,
	}

	return c.presenter.SendTalkToNPCResult(ctx, sessionID, resp)
}
