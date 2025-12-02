package usecaseadapter

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// ConsumeUseCaseAdapter 实现 ConsumeUseCase 接口（用于 SkillSys 依赖）
type ConsumeUseCaseAdapter struct{}

// NewConsumeUseCaseAdapter 创建 ConsumeUseCase 适配器
func NewConsumeUseCaseAdapter() interfaces.ConsumeUseCase {
	return &ConsumeUseCaseAdapter{}
}

// CheckConsume 检查消耗是否足够
func (a *ConsumeUseCaseAdapter) CheckConsume(ctx context.Context, roleID uint64, items []*jsonconf.ItemAmount) error {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error: %v", err)
		return err
	}
	return playerRole.CheckConsume(ctx, items)
}

// ApplyConsume 应用消耗（扣除货币、物品等）
func (a *ConsumeUseCaseAdapter) ApplyConsume(ctx context.Context, roleID uint64, items []*jsonconf.ItemAmount) error {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error: %v", err)
		return err
	}
	return playerRole.ApplyConsume(ctx, items)
}
