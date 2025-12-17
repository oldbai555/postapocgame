package skill

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/fuben"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/app/playeractor/service/consume"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// SkillController 技能控制器
type SkillController struct {
	deps                Deps
	learnSkillUseCase   *LearnSkillUseCase
	upgradeSkillUseCase *UpgradeSkillUseCase
	presenter           *SkillPresenter
}

// NewSkillController 创建技能控制器
func NewSkillController(rt *runtime.Runtime) *SkillController {
	d := depsFromRuntime(rt)
	consumeUseCase := consume.NewConsumeUseCase(d.PlayerRepo, d.EventPublisher)

	return &SkillController{
		deps:                d,
		learnSkillUseCase:   NewLearnSkillUseCase(d, consumeUseCase),
		upgradeSkillUseCase: NewUpgradeSkillUseCase(d, consumeUseCase),
		presenter:           NewSkillPresenter(d.NetworkGateway),
	}
}

// HandleLearnSkill 处理学习技能请求
func (c *SkillController) HandleLearnSkill(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	skillSys := GetSkillSys(ctx)
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
	}

	// 发送响应
	return c.presenter.SendLearnSkillResult(ctx, sessionID, resp)
}

// HandleUpgradeSkill 处理升级技能请求
func (c *SkillController) HandleUpgradeSkill(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	skillSys := GetSkillSys(ctx)
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
	fubenSys := fuben.GetFubenSys(ctx)
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

// init 注册协议
// init() 函数已移除，注册逻辑迁移至 playeractor/register.RegisterAll()
