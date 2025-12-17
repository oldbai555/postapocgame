package bag

import "context"

// HasItemUseCase 检查物品用例（Phase2B：小 service，内部使用 Deps 聚合依赖）。
type HasItemUseCase struct {
	deps Deps
}

// NewHasItemUseCase 创建检查物品用例。
func NewHasItemUseCase(deps Deps) *HasItemUseCase {
	return &HasItemUseCase{deps: deps}
}

// Execute 执行检查物品用例
func (uc *HasItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) (bool, error) {
	if count == 0 {
		return true, nil
	}

	acc, err := newAccessor(ctx, uc.deps.PlayerRepo)
	if err != nil {
		return false, err
	}
	return acc.totalCount(itemID) >= count, nil
}
