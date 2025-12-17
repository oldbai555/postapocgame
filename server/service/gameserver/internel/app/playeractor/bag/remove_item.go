package bag

import (
	"context"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// RemoveItemUseCase 移除物品用例（Phase2B：小 service，内部使用 Deps 聚合依赖）。
type RemoveItemUseCase struct {
	deps Deps
}

// NewRemoveItemUseCase 创建移除物品用例。
func NewRemoveItemUseCase(deps Deps) *RemoveItemUseCase {
	return &RemoveItemUseCase{deps: deps}
}

// Execute 执行移除物品用例
func (uc *RemoveItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) error {
	if count == 0 {
		return nil
	}

	acc, err := newAccessor(ctx, uc.deps.PlayerRepo)
	if err != nil {
		return customerr.Wrap(err)
	}
	if err := acc.removeItem(itemID, count); err != nil {
		return err
	}

	// 发布事件
	uc.deps.EventPublisher.PublishPlayerEvent(ctx, gevent.OnItemRemove, map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	})

	return nil
}
