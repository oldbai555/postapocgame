package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/adapter/system"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/skill"
)

// SkillController 技能控制器
type SkillController struct {
	learnSkillUseCase   *skill.LearnSkillUseCase
	upgradeSkillUseCase *skill.UpgradeSkillUseCase
	presenter           *presenter.SkillPresenter
}

// NewSkillController 创建技能控制器
func NewSkillController() *SkillController {
	container := di.GetContainer()
	learnSkillUC := skill.NewLearnSkillUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())
	upgradeSkillUC := skill.NewUpgradeSkillUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())

	// 注入依赖
	levelUseCase := system.NewLevelUseCaseAdapter()
	consumeUseCase := usecaseadapter.NewConsumeUseCaseAdapter()
	learnSkillUC.SetDependencies(levelUseCase, consumeUseCase)
	upgradeSkillUC.SetDependencies(consumeUseCase)

	return &SkillController{
		learnSkillUseCase:   learnSkillUC,
		upgradeSkillUseCase: upgradeSkillUC,
		presenter:           presenter.NewSkillPresenter(container.NetworkGateway()),
	}
}

// HandleLearnSkill 处理学习技能请求
func (c *SkillController) HandleLearnSkill(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SLearnSkillReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 执行学习技能用例
	err = c.learnSkillUseCase.Execute(ctx, roleID, req.SkillId)

	// 构建响应
	resp := &protocol.S2CLearnSkillResultReq{
		Success: err == nil,
		SkillId: req.SkillId,
	}

	if err != nil {
		resp.Message = err.Error()
	} else {
		resp.Message = "学习成功"
		// 触发任务事件（学习技能）
		questSys := system.GetQuestSys(ctx)
		if questSys != nil {
			if err := questSys.UpdateQuestProgressByType(ctx, uint32(protocol.QuestType_QuestTypeLearnSkill), 0, 1); err != nil {
				log.Warnf("Update quest progress failed: %v", err)
			}
		}
	}

	// 发送响应
	return c.presenter.SendLearnSkillResult(ctx, sessionID, resp)
}

// HandleUpgradeSkill 处理升级技能请求
func (c *SkillController) HandleUpgradeSkill(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SUpgradeSkillReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 执行升级技能用例
	skillLevel, err := c.upgradeSkillUseCase.Execute(ctx, roleID, req.SkillId)

	// 构建响应
	resp := &protocol.S2CUpgradeSkillResultReq{
		Success:    err == nil,
		SkillId:    req.SkillId,
		SkillLevel: skillLevel,
	}

	if err != nil {
		resp.Message = err.Error()
	} else {
		resp.Message = "升级成功"
	}

	// 发送响应
	return c.presenter.SendUpgradeSkillResult(ctx, sessionID, resp)
}
