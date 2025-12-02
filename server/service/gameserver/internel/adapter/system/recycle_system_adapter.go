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
