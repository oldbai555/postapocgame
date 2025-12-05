package controller

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	system2 "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/quest"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/reward"
	"postapocgame/server/service/gameserver/internel/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
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
	acceptQuestUC := quest.NewAcceptQuestUseCase(deps.PlayerGateway(), deps.ConfigGateway())
	updateProgressUC := quest.NewUpdateQuestProgressUseCase(deps.PlayerGateway(), deps.ConfigGateway())
	submitQuestUC := quest.NewSubmitQuestUseCase(deps.PlayerGateway(), deps.ConfigGateway())

	// 注入依赖
	levelUseCase := system2.NewLevelUseCaseAdapter()
	rewardUseCase := reward.NewRewardUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway())
	acceptQuestUC.SetDependencies(levelUseCase)
	submitQuestUC.SetDependencies(levelUseCase, rewardUseCase)

	return &QuestController{
		acceptQuestUseCase:    acceptQuestUC,
		updateProgressUseCase: updateProgressUC,
		submitQuestUseCase:    submitQuestUC,
		presenter:             presenter.NewQuestPresenter(deps.NetworkGateway()),
	}
}

// HandleTalkToNPC 处理NPC对话请求
func (c *QuestController) HandleTalkToNPC(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	questSys := system2.GetQuestSys(ctx)
	if questSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "任务系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2STalkToNPCReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 检查NPC配置
	npcConfigRaw := deps.ConfigGateway().GetNPCSceneConfig(req.NpcId)
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
