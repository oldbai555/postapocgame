package gateway

import (
	"context"
	"postapocgame/server/internal/database"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// BlacklistRepositoryAdapter 实现黑名单数据访问
type BlacklistRepositoryAdapter struct{}

// NewBlacklistRepositoryAdapter 创建黑名单仓储适配器
func NewBlacklistRepositoryAdapter() interfaces.BlacklistRepository {
	return &BlacklistRepositoryAdapter{}
}

// Add 添加黑名单
func (a *BlacklistRepositoryAdapter) Add(_ context.Context, ownerID uint64, targetID uint64, reason string) error {
	return database.AddToBlacklist(targetID, ownerID, reason)
}

// Remove 移除黑名单
func (a *BlacklistRepositoryAdapter) Remove(_ context.Context, ownerID uint64, targetID uint64) error {
	return database.RemoveFromBlacklist(targetID, ownerID)
}

// GetList 获取黑名单列表
func (a *BlacklistRepositoryAdapter) GetList(_ context.Context, ownerID uint64) ([]uint64, error) {
	items, err := database.GetBlacklist(ownerID)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ids = append(ids, item.RoleId)
	}
	return ids, nil
}

// IsBlocked 检查是否在黑名单
func (a *BlacklistRepositoryAdapter) IsBlocked(_ context.Context, ownerID uint64, targetID uint64) (bool, error) {
	return database.IsInBlacklist(targetID, ownerID)
}
