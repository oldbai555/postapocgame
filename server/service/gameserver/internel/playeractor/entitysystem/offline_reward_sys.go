package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

const (
	// 离线收益上限（默认24小时）
	maxOfflineSeconds = 24 * 60 * 60
	// 每分钟收益基础值（经验、金币等）
	baseRewardPerMinute = 10
)

// OfflineRewardSys 离线收益系统
type OfflineRewardSys struct {
	*BaseSystem
	offlineData *protocol.SiOfflineRewardData
}

// NewOfflineRewardSys 创建离线收益系统
func NewOfflineRewardSys() *OfflineRewardSys {
	sys := &OfflineRewardSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysOfflineReward)),
	}
	return sys
}

func GetOfflineRewardSys(ctx context.Context) *OfflineRewardSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysOfflineReward))
	if system == nil {
		log.Errorf("not found system [%v] error:%v", protocol.SystemId_SysOfflineReward, err)
		return nil
	}
	sys := system.(*OfflineRewardSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error:%v", protocol.SystemId_SysOfflineReward, err)
		return nil
	}
	return sys
}

// OnInit 系统初始化
func (ors *OfflineRewardSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("offline reward sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果offline_reward_data不存在，则初始化
	if binaryData.OfflineRewardData == nil {
		binaryData.OfflineRewardData = &protocol.SiOfflineRewardData{
			LastLogoutTime:      0,
			TotalOfflineSeconds: 0,
			RewardClaimed:       false,
			LastClaimTime:       0,
		}
	}
	ors.offlineData = binaryData.OfflineRewardData
}

// OnRoleLogout 登出回调（记录登出时间）
func (ors *OfflineRewardSys) OnRoleLogout(ctx context.Context) {
	if ors.offlineData == nil {
		return
	}

	// 记录登出时间
	ors.offlineData.LastLogoutTime = servertime.Now().Unix()
	ors.offlineData.RewardClaimed = false

	log.Infof("[OfflineRewardSys] OnRoleLogout: LastLogoutTime=%d", ors.offlineData.LastLogoutTime)
}

// OnRoleLogin 登录回调（计算离线时间）
func (ors *OfflineRewardSys) OnRoleLogin(ctx context.Context) {
	if ors.offlineData == nil {
		return
	}

	// 如果有登出时间，计算离线时间
	if ors.offlineData.LastLogoutTime > 0 {
		now := servertime.Now().Unix()
		offlineSeconds := now - ors.offlineData.LastLogoutTime

		// 限制离线时间上限
		if offlineSeconds > maxOfflineSeconds {
			offlineSeconds = maxOfflineSeconds
		}

		// 累计离线时间
		ors.offlineData.TotalOfflineSeconds = offlineSeconds

		log.Infof("[OfflineRewardSys] OnRoleLogin: OfflineSeconds=%d", offlineSeconds)
	}
}

// GetOfflineSeconds 获取离线时间（秒）
func (ors *OfflineRewardSys) GetOfflineSeconds() int64 {
	if ors.offlineData == nil {
		return 0
	}
	return ors.offlineData.TotalOfflineSeconds
}

// CalculateRewards 计算离线收益
func (ors *OfflineRewardSys) CalculateRewards(ctx context.Context) []*jsonconf.ItemAmount {
	if ors.offlineData == nil || ors.offlineData.TotalOfflineSeconds <= 0 {
		return nil
	}

	// 计算离线分钟数
	offlineMinutes := ors.offlineData.TotalOfflineSeconds / 60
	if offlineMinutes <= 0 {
		return nil
	}

	// 计算收益
	rewards := make([]*jsonconf.ItemAmount, 0)

	// 经验收益：每分钟 baseRewardPerMinute 经验
	expReward := int64(offlineMinutes) * baseRewardPerMinute
	if expReward > 0 {
		rewards = append(rewards, &jsonconf.ItemAmount{
			ItemType: uint32(protocol.ItemType_ItemTypeMoney),
			ItemId:   uint32(protocol.MoneyType_MoneyTypeExp),
			Count:    expReward,
			Bind:     1, // 离线收益默认绑定
		})
	}

	// 金币收益：每分钟 baseRewardPerMinute 金币
	goldReward := int64(offlineMinutes) * baseRewardPerMinute
	if goldReward > 0 {
		rewards = append(rewards, &jsonconf.ItemAmount{
			ItemType: uint32(protocol.ItemType_ItemTypeMoney),
			ItemId:   uint32(protocol.MoneyType_MoneyTypeGoldCoin),
			Count:    goldReward,
			Bind:     1, // 离线收益默认绑定
		})
	}

	return rewards
}

// ClaimReward 领取离线收益
func (ors *OfflineRewardSys) ClaimReward(ctx context.Context) ([]*jsonconf.ItemAmount, error) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return nil, err
	}

	if ors.offlineData == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "offline reward data not initialized")
	}

	// 检查是否已领取
	if ors.offlineData.RewardClaimed {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "offline reward already claimed")
	}

	// 检查是否有离线收益
	if ors.offlineData.TotalOfflineSeconds <= 0 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "no offline reward to claim")
	}

	// 计算收益
	rewards := ors.CalculateRewards(ctx)
	if len(rewards) == 0 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "no offline reward to claim")
	}

	// 发放奖励
	if err := playerRole.GrantRewards(ctx, rewards); err != nil {
		log.Errorf("grant offline reward failed: %v", err)
		return nil, err
	}

	// 标记已领取
	ors.offlineData.RewardClaimed = true
	ors.offlineData.LastClaimTime = servertime.Now().Unix()
	ors.offlineData.TotalOfflineSeconds = 0 // 清零离线时间

	log.Infof("[OfflineRewardSys] ClaimReward: Rewards=%d", len(rewards))

	return rewards, nil
}

// handleClaimOfflineReward 处理领取离线收益请求
func handleClaimOfflineReward(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SClaimOfflineRewardReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		resp := &protocol.S2CClaimOfflineRewardResultReq{
			Success: false,
			Message: "玩家角色不存在",
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CClaimOfflineRewardResult), resp)
	}

	// 获取离线收益系统
	roleCtx := playerRole.WithContext(ctx)
	offlineRewardSys := GetOfflineRewardSys(roleCtx)
	if offlineRewardSys == nil {
		resp := &protocol.S2CClaimOfflineRewardResultReq{
			Success: false,
			Message: "离线收益系统未初始化",
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CClaimOfflineRewardResult), resp)
	}

	// 获取离线时间
	offlineSeconds := offlineRewardSys.GetOfflineSeconds()

	// 领取收益
	rewards, err := offlineRewardSys.ClaimReward(roleCtx)

	// 构造响应
	resp := &protocol.S2CClaimOfflineRewardResultReq{
		Success:        err == nil,
		OfflineSeconds: offlineSeconds,
		ClaimedTime:    servertime.Now().Unix(),
	}

	if err != nil {
		resp.Message = err.Error()
		resp.Success = false
	} else {
		resp.Message = "领取成功"
		// 转换奖励列表
		if rewards != nil && len(rewards) > 0 {
			resp.Rewards = make([]*protocol.ItemAmount, 0, len(rewards))
			for _, reward := range rewards {
				resp.Rewards = append(resp.Rewards, &protocol.ItemAmount{
					ItemType: reward.ItemType,
					ItemId:   reward.ItemId,
					Count:    reward.Count,
					Bind:     reward.Bind,
				})
			}
		}
		// 推送背包和货币数据更新
		pushBagData(roleCtx, sessionId)
		pushMoneyData(roleCtx, sessionId)
	}

	// 发送响应
	if sendErr := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CClaimOfflineRewardResult), resp); sendErr != nil {
		return sendErr
	}

	return err
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysOfflineReward), func() iface.ISystem {
		return NewOfflineRewardSys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SClaimOfflineReward), handleClaimOfflineReward)
	})
}
