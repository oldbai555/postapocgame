package interfaces

import "context"

// LevelUseCase 等级系统用例接口（Use Case 层定义，用于 ItemUseSys、SkillSys 依赖）
type LevelUseCase interface {
	// AddExp 添加经验值
	AddExp(ctx context.Context, roleID uint64, exp uint64) error

	// GetLevel 获取当前等级
	GetLevel(ctx context.Context, roleID uint64) (uint32, error)
}
