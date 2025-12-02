package interfaces

import "context"

// DailyActivityUseCase 日常活跃度系统用例接口（Use Case 层定义，用于 QuestSys 依赖）
type DailyActivityUseCase interface {
	// AddActivePoints 添加活跃点
	AddActivePoints(ctx context.Context, roleID uint64, points uint32) error
}
