package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/di"
	recycleusecase "postapocgame/server/service/gameserver/internel/usecase/recycle"
	"sync"
)

// RecycleSystemAdapter 回收系统适配器
//
// 生命周期职责：
// - 暂未实现生命周期方法（回收系统以单例形式暴露，不依赖 Actor 生命周期）
//
// 业务逻辑：所有业务逻辑（回收物品校验、奖励计算）均在 UseCase 层实现
// 状态管理：以单例形式管理，通过 sync.Once 确保只初始化一次
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type RecycleSystemAdapter struct {
	recycleItemUseCase *recycleusecase.RecycleItemUseCase
}

var (
	recycleAdapter     *RecycleSystemAdapter
	recycleAdapterOnce sync.Once
)

// NewRecycleSystemAdapter 创建回收系统适配器
func NewRecycleSystemAdapter() *RecycleSystemAdapter {
	container := di.GetContainer()
	recycleUC := recycleusecase.NewRecycleItemUseCase(container.ConfigGateway())
	recycleUC.SetDependencies(NewBagUseCaseAdapter(), usecaseadapter.NewRewardUseCaseAdapter())
	return &RecycleSystemAdapter{
		recycleItemUseCase: recycleUC,
	}
}

// getRecycleSysInstance 获取回收系统适配器实例
func getRecycleSysInstance() *RecycleSystemAdapter {
	recycleAdapterOnce.Do(func() {
		recycleAdapter = NewRecycleSystemAdapter()
	})
	return recycleAdapter
}

// RecycleItem 回收物品
func (a *RecycleSystemAdapter) RecycleItem(ctx context.Context, roleID uint64, itemID uint32, count uint32) ([]*protocol.ItemAmount, error) {
	return a.recycleItemUseCase.Execute(ctx, roleID, itemID, count)
}
