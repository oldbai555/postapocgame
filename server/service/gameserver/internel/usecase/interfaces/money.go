package interfaces

import "context"

// MoneyUseCase 货币系统用例接口（Use Case 层定义，用于 LevelSys 依赖）
// 注意：此接口在 MoneySys 重构后会被实现
type MoneyUseCase interface {
	// UpdateExp 更新经验值（经验作为货币的一种）
	UpdateExp(ctx context.Context, roleID uint64, exp int64) error

	// GetExp 获取经验值
	GetExp(ctx context.Context, roleID uint64) (int64, error)
}
