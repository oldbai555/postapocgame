package friend

import (
	"context"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// BlacklistUseCase 黑名单相关操作
type BlacklistUseCase struct {
	repo interfaces.BlacklistRepository
}

// NewBlacklistUseCase 创建用例
func NewBlacklistUseCase(repo interfaces.BlacklistRepository) *BlacklistUseCase {
	return &BlacklistUseCase{repo: repo}
}

// Add 添加到黑名单
func (uc *BlacklistUseCase) Add(ctx context.Context, ownerID uint64, targetID uint64, reason string) error {
	if ownerID == 0 || targetID == 0 {
		return customerr.NewError("参数错误")
	}
	if ownerID == targetID {
		return customerr.NewError("不能拉黑自己")
	}
	return uc.repo.Add(ctx, ownerID, targetID, reason)
}

// Remove 从黑名单移除
func (uc *BlacklistUseCase) Remove(ctx context.Context, ownerID uint64, targetID uint64) error {
	if ownerID == 0 || targetID == 0 {
		return customerr.NewError("参数错误")
	}
	return uc.repo.Remove(ctx, ownerID, targetID)
}

// Query 查询黑名单
func (uc *BlacklistUseCase) Query(ctx context.Context, ownerID uint64) ([]uint64, error) {
	if ownerID == 0 {
		return nil, customerr.NewError("未登录")
	}
	return uc.repo.GetList(ctx, ownerID)
}
