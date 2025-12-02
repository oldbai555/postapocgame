package dailyactivity

import (
	"context"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	domain "postapocgame/server/service/gameserver/internel/domain/dailyactivity"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// ClaimRewardUseCase 领取活跃奖励
type ClaimRewardUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
	rewardUseCase interfaces.RewardUseCase
}

func NewClaimRewardUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
	rewardUseCase interfaces.RewardUseCase,
) *ClaimRewardUseCase {
	return &ClaimRewardUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
		rewardUseCase: rewardUseCase,
	}
}

func (uc *ClaimRewardUseCase) Execute(ctx context.Context, roleID uint64, rewardID uint32) error {
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}
	data := domain.EnsureData(binaryData)
	if data == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily activity data not initialized")
	}

	var state *protocol.DailyActivityRewardState
	for _, st := range data.RewardStates {
		if st != nil && st.RewardId == rewardID {
			state = st
			break
		}
	}
	if state == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "reward config not found: %d", rewardID)
	}
	if state.Claimed {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "reward already claimed")
	}

	rawCfg, ok := uc.configManager.GetDailyActivityRewardConfig(rewardID)
	if !ok || rawCfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "reward config missing: %d", rewardID)
	}
	cfg, ok := rawCfg.(*jsonconf.DailyActivityRewardConfig)
	if !ok || cfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid reward config type: %d", rewardID)
	}
	if data.TodayPoints < cfg.RequiredPoint {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not enough active points to claim reward")
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
		if err := uc.rewardUseCase.GrantRewards(ctx, roleID, rewards); err != nil {
			return customerr.Wrap(err)
		}
	}

	state.Claimed = true
	return nil
}
