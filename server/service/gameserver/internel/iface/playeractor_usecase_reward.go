package iface

import (
	"context"
	"postapocgame/server/internal/jsonconf"
)

// RewardUseCase 奖励发放用例接口（Use Case 层定义）
// 注意：此接口在后续系统重构后会被实现
type RewardUseCase interface {
	// GrantRewards 发放奖励
	GrantRewards(ctx context.Context, roleID uint64, rewards []*jsonconf.ItemAmount) error
}
