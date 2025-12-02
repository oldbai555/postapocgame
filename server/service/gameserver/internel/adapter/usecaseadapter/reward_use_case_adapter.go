package usecaseadapter

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// RewardUseCaseAdapter 实现 RewardUseCase 接口（用于 FubenSys 依赖）
type RewardUseCaseAdapter struct{}

// NewRewardUseCaseAdapter 创建 RewardUseCase 适配器
func NewRewardUseCaseAdapter() interfaces.RewardUseCase {
	return &RewardUseCaseAdapter{}
}

// GrantRewards 发放奖励
func (a *RewardUseCaseAdapter) GrantRewards(ctx context.Context, roleID uint64, rewards []*jsonconf.ItemAmount) error {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error: %v", err)
		return err
	}
	return playerRole.GrantRewards(ctx, rewards)
}
