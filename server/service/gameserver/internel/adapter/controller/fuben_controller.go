package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/adapter/system"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/fuben"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// FubenController 副本控制器
type FubenController struct {
	enterDungeonUseCase  *fuben.EnterDungeonUseCase
	settleDungeonUseCase *fuben.SettleDungeonUseCase
	presenter            *presenter.FubenPresenter
	dungeonGateway       interfaces.DungeonServerGateway
}

// NewFubenController 创建副本控制器
func NewFubenController() *FubenController {
	container := di.GetContainer()
	enterDungeonUC := fuben.NewEnterDungeonUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())
	settleDungeonUC := fuben.NewSettleDungeonUseCase(container.PlayerGateway())

	// 注入依赖
	consumeUseCase := usecaseadapter.NewConsumeUseCaseAdapter()
	levelUseCase := system.NewLevelUseCaseAdapter()
	rewardUseCase := usecaseadapter.NewRewardUseCaseAdapter()
	enterDungeonUC.SetDependencies(consumeUseCase)
	settleDungeonUC.SetDependencies(levelUseCase, rewardUseCase)

	return &FubenController{
		enterDungeonUseCase:  enterDungeonUC,
		settleDungeonUseCase: settleDungeonUC,
		presenter:            presenter.NewFubenPresenter(container.NetworkGateway()),
		dungeonGateway:       container.DungeonServerGateway(),
	}
}

// HandleEnterDungeon 处理进入副本请求
func (c *FubenController) HandleEnterDungeon(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	fubenSys := system.GetFubenSys(ctx)
	if fubenSys == nil {
		sessionID, _ := adaptercontext.GetSessionIDFromContext(ctx)
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "副本系统未开启",
			DungeonId: 0,
		}
		return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
	}

	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SEnterDungeonReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
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

	// 汇总属性
	var syncAttrData *protocol.SyncAttrData
	if playerRole != nil {
		attrCalcRaw := playerRole.GetAttrCalculator()
		if attrCalc, ok := attrCalcRaw.(interfaces.IAttrCalculator); ok && attrCalc != nil {
			allAttrs := attrCalc.CalculateAllAttrs(ctx)
			if len(allAttrs) > 0 {
				syncAttrData = &protocol.SyncAttrData{
					AttrData: allAttrs,
				}
			}
		}
	}

	// 获取技能列表
	skillSys := system.GetSkillSys(ctx)
	var skillMap map[uint32]uint32
	if skillSys != nil {
		skillMap, _ = skillSys.GetSkillMap(ctx)
	} else {
		skillMap = make(map[uint32]uint32)
	}

	// 构造进入副本请求
	reqData, err := internal.Marshal(&protocol.G2DEnterDungeonReq{
		SessionId:    sessionID,
		PlatformId:   gshare.GetPlatformId(),
		SrvId:        gshare.GetSrvId(),
		SimpleData:   roleInfo,
		SyncAttrData: syncAttrData,
		SkillMap:     skillMap,
		DungeonId:    req.DungeonId,
		Difficulty:   req.Difficulty,
	})
	if err != nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "系统错误",
			DungeonId: req.DungeonId,
		}
		return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
	}

	// 使用带SessionId的异步RPC调用
	srvType := uint8(protocol.SrvType_SrvTypeDungeonServer)
	err = c.dungeonGateway.AsyncCall(ctx, srvType, sessionID, uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), reqData)
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
	fubenSys := system.GetFubenSys(ctx)
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
