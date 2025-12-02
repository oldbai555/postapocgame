package system

import (
	"context"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// dailyActivityUseCaseAdapter 实现 interfaces.DailyActivityUseCase
type dailyActivityUseCaseAdapter struct{}

// NewDailyActivityUseCaseAdapter 创建 DailyActivityUseCase 适配器
func NewDailyActivityUseCaseAdapter() interfaces.DailyActivityUseCase {
	return &dailyActivityUseCaseAdapter{}
}

// AddActivePoints 添加活跃点
func (a *dailyActivityUseCaseAdapter) AddActivePoints(ctx context.Context, roleID uint64, points uint32) error {
	sys := GetDailyActivitySys(ctx)
	if sys == nil {
		log.Warnf("DailyActivitySys not ready: RoleID=%d, Points=%d", roleID, points)
		return nil
	}
	return sys.AddActivePoints(ctx, points)
}
