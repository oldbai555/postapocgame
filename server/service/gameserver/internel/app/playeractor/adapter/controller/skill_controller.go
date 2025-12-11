package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/router"
	system2 "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/consume"
	skill2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/skill"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// SkillController 技能控制器
type SkillController struct {
	learnSkillUseCase   *skill2.LearnSkillUseCase
	upgradeSkillUseCase *skill2.UpgradeSkillUseCase
	presenter           *presenter.SkillPresenter
}

// NewSkillController 创建技能控制器
func NewSkillController() *SkillController {
	learnSkillUC := skill2.NewLearnSkillUseCase(deps.PlayerGateway(), deps.ConfigGateway(), deps.DungeonServerGateway())
	upgradeSkillUC := skill2.NewUpgradeSkillUseCase(deps.PlayerGateway(), deps.ConfigGateway(), deps.DungeonServerGateway())

	// 注入依赖
	levelUseCase := system2.NewLevelUseCaseAdapter()
	consumeUseCase := consume.NewConsumeUseCase(deps.PlayerGateway(), deps.EventPublisher())
	learnSkillUC.SetDependencies(levelUseCase, consumeUseCase)
	upgradeSkillUC.SetDependencies(consumeUseCase)

	return &SkillController{
		learnSkillUseCase:   learnSkillUC,
		upgradeSkillUseCase: upgradeSkillUC,
		presenter:           presenter.NewSkillPresenter(deps.NetworkGateway()),
	}
}

// HandleLearnSkill 处理学习技能请求
func (c *SkillController) HandleLearnSkill(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	skillSys := system2.GetSkillSys(ctx)
	if skillSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "技能系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SLearnSkillReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
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
		questSys := system2.GetQuestSys(ctx)
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
	// 检查系统是否开启
	skillSys := system2.GetSkillSys(ctx)
	if skillSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "技能系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SUpgradeSkillReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
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

// HandleUseSkill 处理战斗服内技能释放请求（转发给 DungeonActor）
// 说明：技能的实际判定与伤害计算仍在 DungeonActor 的 FightSys 中执行。
func (c *SkillController) HandleUseSkill(ctx context.Context, msg *network.ClientMessage) error {
	// 检查副本/战斗相关系统是否开启（这里复用 FubenSys 的开关做最小保护）
	fubenSys := system2.GetFubenSys(ctx)
	if fubenSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "副本系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SUseSkill payload")
	}

	// 直接复用客户端上报的 Proto 数据，交由 DungeonActor 解析处理
	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdUseSkill), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		skillController := NewSkillController()
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SLearnSkill), skillController.HandleLearnSkill)
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUpgradeSkill), skillController.HandleUpgradeSkill)
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUseSkill), skillController.HandleUseSkill)
	})
}
