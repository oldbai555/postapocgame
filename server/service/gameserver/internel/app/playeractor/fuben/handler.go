package fuben

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// FubenController 副本控制器
type FubenController struct {
	enterDungeonUseCase  *EnterDungeonUseCase
	settleDungeonUseCase *SettleDungeonUseCase
	presenter            *FubenPresenter
}

// NewFubenController 创建副本控制器
func NewFubenController(rt *runtime.Runtime) *FubenController {
	d := depsFromRuntime(rt)
	return &FubenController{
		enterDungeonUseCase:  NewEnterDungeonUseCase(d),
		settleDungeonUseCase: NewSettleDungeonUseCase(d),
		presenter:            NewFubenPresenter(d.NetworkGateway),
	}
}

// HandleEnterDungeon 处理进入副本请求
func (c *FubenController) HandleEnterDungeon(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	fubenSys := GetFubenSys(ctx)
	if fubenSys == nil {
		sessionID, _ := gshare.GetSessionIDFromContext(ctx)
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "副本系统未开启",
			DungeonId: 0,
		}
		return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SEnterDungeonReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		return err
	}

	// 执行进入副本用例
	err = c.enterDungeonUseCase.Execute(ctx, roleID, req.DungeonId, req.Difficulty)
	if err != nil {
		// 构建错误响应
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   err.Error(),
			DungeonId: req.DungeonId,
		}
		return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
	}

	// 获取角色信息
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "角色信息不存在",
			DungeonId: req.DungeonId,
		}
		return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
	}

	// 属性系统已移除，这里不再汇总属性；DungeonActor 可按自身规则处理属性

	// 构造进入副本请求
	reqData, err := internal.Marshal(&protocol.G2DEnterDungeonReq{
		SessionId:  sessionID,
		PlatformId: gshare.GetPlatformId(),
		SrvId:      gshare.GetSrvId(),
		SimpleData: roleInfo,
		// SyncAttrData 与 SkillMap 已从此处移除，DungeonActor 可自行获取所需数据
		DungeonId:  req.DungeonId,
		Difficulty: req.Difficulty,
	})
	if err != nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "系统错误",
			DungeonId: req.DungeonId,
		}
		return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
	}

	// 使用带SessionId的异步调用（通过 DungeonActorMsgId 枚举）
	fubenSys2 := GetFubenSys(ctx)
	if fubenSys2 == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "副本系统未开启")
	}
	err = fubenSys2.deps.DungeonGateway.AsyncCall(ctx, sessionID, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdEnterDungeon), reqData)
	if err != nil {
		log.Errorf("call dungeon service enter dungeon failed: %v", err)
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "进入副本失败",
			DungeonId: req.DungeonId,
		}
		return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
	}

	// 发送成功响应
	resp := &protocol.S2CEnterDungeonResultReq{
		Success:   true,
		Message:   "进入副本成功",
		DungeonId: req.DungeonId,
	}
	return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
}

// HandleSettleDungeon 处理副本结算的RPC请求
func (c *FubenController) HandleSettleDungeon(ctx context.Context, sessionID string, data []byte) error {
	// 检查系统是否开启
	fubenSys := GetFubenSys(ctx)
	if fubenSys == nil {
		log.Errorf("fuben system not enabled")
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "副本系统未开启")
	}

	var req protocol.D2GSettleDungeonReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal settle dungeon request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("received dungeon settlement: RoleId=%d, DungeonID=%d, Success=%v, Rewards=%d",
		req.RoleId, req.DungeonId, req.Success, len(req.Rewards))

	// 将协议层 RewardItem 转换为 jsonconf.DungeonReward（供结算用例使用）
	rewards := make([]*jsonconf.DungeonReward, 0, len(req.Rewards))
	for _, r := range req.Rewards {
		if r == nil {
			continue
		}
		rewards = append(rewards, &jsonconf.DungeonReward{
			Type:   r.Type,
			ItemID: r.ItemId,
			Count:  r.Count,
			Rate:   1, // 结算结果已确定，这里视为 100% 发放
		})
	}

	// 执行副本结算用例
	err := c.settleDungeonUseCase.Execute(ctx, req.RoleId, req.DungeonId, req.Difficulty, req.Success, rewards)
	if err != nil {
		log.Errorf("settle dungeon failed: %v", err)
		return err
	}

	return nil
}

// HandleEnterDungeonSuccess 处理进入副本成功通知
func (c *FubenController) HandleEnterDungeonSuccess(ctx context.Context, sessionID string, data []byte) error {
	var req protocol.D2GEnterDungeonSuccessReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal enter dungeon success request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("role entered dungeon successfully: RoleId=%d, SessionId=%s", req.RoleId, req.SessionId)
	// 这里可以添加后续处理逻辑，比如更新玩家状态等
	return nil
}

// init() 函数已移除，注册逻辑迁移至 playeractor/register.RegisterAll()
