package interfaces

import (
	"context"
	"postapocgame/server/internal/jsonconf"
)

// ConsumeUseCase 消耗用例接口（Use Case 层定义）
// 用于检查和应用消耗（货币、物品等）
type ConsumeUseCase interface {
	// CheckConsume 检查消耗是否足够
	CheckConsume(ctx context.Context, roleID uint64, items []*jsonconf.ItemAmount) error

	// ApplyConsume 应用消耗（扣除货币、物品等）
	ApplyConsume(ctx context.Context, roleID uint64, items []*jsonconf.ItemAmount) error
}
