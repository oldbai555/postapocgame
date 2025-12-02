package interfaces

import "context"

// BlacklistRepository 黑名单数据访问接口
type BlacklistRepository interface {
	Add(ctx context.Context, ownerID uint64, targetID uint64, reason string) error
	Remove(ctx context.Context, ownerID uint64, targetID uint64) error
	GetList(ctx context.Context, ownerID uint64) ([]uint64, error)
	IsBlocked(ctx context.Context, ownerID uint64, targetID uint64) (bool, error)
}
