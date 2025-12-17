package level

import (
	"context"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

// levelUseCaseAdapter 实现 iface.LevelUseCase，供其他系统依赖
type levelUseCaseAdapter struct{}

// NewLevelUseCaseAdapter 创建 LevelUseCase 适配器
func NewLevelUseCaseAdapter() iface.LevelUseCase {
	return &levelUseCaseAdapter{}
}

// AddExp 添加经验值
func (a *levelUseCaseAdapter) AddExp(ctx context.Context, roleID uint64, exp uint64) error {
	levelSys := GetLevelSys(ctx)
	if levelSys == nil {
		log.Warnf("LevelSys not found: RoleID=%d", roleID)
		return nil
	}
	return levelSys.AddExp(ctx, exp)
}

// GetLevel 获取当前等级
func (a *levelUseCaseAdapter) GetLevel(ctx context.Context, roleID uint64) (uint32, error) {
	levelSys := GetLevelSys(ctx)
	if levelSys == nil {
		log.Warnf("LevelSys not found: RoleID=%d", roleID)
		return 0, nil
	}
	return levelSys.GetLevel(ctx)
}
