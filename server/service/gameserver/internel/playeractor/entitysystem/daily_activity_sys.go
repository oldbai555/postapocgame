package entitysystem

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"time"
)

// DailyActivitySys 日常活跃系统
type DailyActivitySys struct {
	*BaseSystem
	activityData *protocol.SiDailyActivityData
}

// NewDailyActivitySys 创建日常活跃系统
func NewDailyActivitySys() *DailyActivitySys {
	return &DailyActivitySys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysDailyActivity)),
	}
}

func GetDailyActivitySys(ctx context.Context) *DailyActivitySys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysDailyActivity))
	if system == nil {
		return nil
	}
	activitySys, ok := system.(*DailyActivitySys)
	if !ok || !activitySys.IsOpened() {
		return nil
	}
	return activitySys
}

func (das *DailyActivitySys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("daily activity OnInit get role err:%v", err)
		return
	}

	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	if binaryData.DailyActivityData == nil {
		binaryData.DailyActivityData = &protocol.SiDailyActivityData{
			RewardStates: make([]*protocol.DailyActivityRewardState, 0),
		}
	}
	das.activityData = binaryData.DailyActivityData
	if das.activityData.RewardStates == nil {
		das.activityData.RewardStates = make([]*protocol.DailyActivityRewardState, 0)
	}

	das.ensureRewardStates()
	das.resetIfNeeded(servertime.Now())
	das.syncMoneyBalance(ctx)
}

func (das *DailyActivitySys) GetActivityData() *protocol.SiDailyActivityData {
	return das.activityData
}

// AddActivePoints 增加活跃点
func (das *DailyActivitySys) AddActivePoints(ctx context.Context, amount uint32) error {
	if amount == 0 {
		return nil
	}
	if das.activityData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "activity data not initialized")
	}

	das.resetIfNeeded(servertime.Now())

	das.activityData.Balance += int64(amount)
	das.activityData.TodayPoints += amount

	if err := das.syncMoneyBalance(ctx); err != nil {
		return err
	}

	if playerRole, err := GetIPlayerRoleByContext(ctx); err == nil {
		log.Infof("Active points added: RoleID=%d, Amount=%d, Balance=%d", playerRole.GetPlayerRoleId(), amount, das.activityData.Balance)
	}
	return nil
}

// CostActivePoints 扣除活跃点
func (das *DailyActivitySys) CostActivePoints(ctx context.Context, amount uint32) error {
	if amount == 0 {
		return nil
	}
	if das.activityData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "activity data not initialized")
	}
	if das.activityData.Balance < int64(amount) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "active points not enough")
	}

	das.activityData.Balance -= int64(amount)
	return das.syncMoneyBalance(ctx)
}

// ClaimReward 领取活跃奖励
func (das *DailyActivitySys) ClaimReward(ctx context.Context, rewardId uint32) error {
	if das.activityData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "activity data not initialized")
	}
	state := das.getRewardState(rewardId)
	if state == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "reward config not found: %d", rewardId)
	}
	if state.Claimed {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "reward already claimed")
	}

	cfg, ok := jsonconf.GetConfigManager().GetDailyActivityRewardConfig(rewardId)
	if !ok || cfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "reward config missing: %d", rewardId)
	}
	if das.activityData.TodayPoints < cfg.RequiredPoint {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not enough active points to claim reward")
	}

	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	rewards := make([]*jsonconf.ItemAmount, 0, len(cfg.Rewards)+len(cfg.ExtraItems))
	for _, item := range cfg.Rewards {
		rewards = append(rewards, &jsonconf.ItemAmount{
			ItemType: uint32(item.Type),
			ItemId:   item.ItemId,
			Count:    int64(item.Count),
			Bind:     1,
		})
	}
	for _, item := range cfg.ExtraItems {
		rewards = append(rewards, &jsonconf.ItemAmount{
			ItemType: uint32(item.Type),
			ItemId:   item.ItemId,
			Count:    int64(item.Count),
			Bind:     1,
		})
	}

	if len(rewards) > 0 {
		if err := playerRole.GrantRewards(ctx, rewards); err != nil {
			return err
		}
	}

	state.Claimed = true
	log.Infof("Daily activity reward claimed: RoleID=%d, RewardID=%d", playerRole.GetPlayerRoleId(), rewardId)
	return nil
}

func (das *DailyActivitySys) OnRoleLogin(ctx context.Context) {
	das.resetIfNeeded(servertime.Now())
}

func (das *DailyActivitySys) OnNewDay(ctx context.Context) {
	das.resetForNewDay(servertime.Now())
}

func (das *DailyActivitySys) ensureRewardStates() {
	if das.activityData == nil {
		return
	}
	existing := make(map[uint32]*protocol.DailyActivityRewardState)
	for _, state := range das.activityData.RewardStates {
		if state != nil {
			existing[state.RewardId] = state
		}
	}

	configs := jsonconf.GetConfigManager().GetAllDailyActivityRewardConfigs()
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		if _, ok := existing[cfg.RewardId]; !ok {
			state := &protocol.DailyActivityRewardState{
				RewardId: cfg.RewardId,
				Claimed:  false,
			}
			das.activityData.RewardStates = append(das.activityData.RewardStates, state)
			existing[cfg.RewardId] = state
		}
	}
}

func (das *DailyActivitySys) getRewardState(rewardId uint32) *protocol.DailyActivityRewardState {
	if das.activityData == nil {
		return nil
	}
	for _, state := range das.activityData.RewardStates {
		if state != nil && state.RewardId == rewardId {
			return state
		}
	}
	return nil
}

func (das *DailyActivitySys) resetIfNeeded(now time.Time) {
	if das.activityData == nil {
		return
	}
	last := time.Unix(das.activityData.LastResetTime, 0).In(time.Local)
	now = now.In(time.Local)
	if das.activityData.LastResetTime == 0 || last.Year() != now.Year() || last.YearDay() != now.YearDay() {
		das.resetForNewDay(now)
	}
}

func (das *DailyActivitySys) resetForNewDay(now time.Time) {
	if das.activityData == nil {
		return
	}
	das.activityData.TodayPoints = 0
	das.activityData.LastResetTime = now.Unix()
	for _, state := range das.activityData.RewardStates {
		if state != nil {
			state.Claimed = false
		}
	}
}

func (das *DailyActivitySys) syncMoneyBalance(ctx context.Context) error {
	if das.activityData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "activity data not initialized")
	}
	moneySys := GetMoneySys(ctx)
	if moneySys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money system not ready")
	}
	current := moneySys.GetAmount(uint32(protocol.MoneyType_MoneyTypeActivePoint))
	delta := das.activityData.Balance - current
	if delta == 0 {
		return nil
	}
	return moneySys.UpdateBalanceTx(ctx, uint32(protocol.MoneyType_MoneyTypeActivePoint), delta, nil)
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysDailyActivity), func() iface.ISystem {
		return NewDailyActivitySys()
	})
}
