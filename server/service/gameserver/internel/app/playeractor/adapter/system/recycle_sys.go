package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	recycleusecase "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/recycle"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/reward"
	"sync"
)

type RecycleSystemAdapter struct {
	recycleItemUseCase *recycleusecase.RecycleItemUseCase
}

var (
	recycleAdapter     *RecycleSystemAdapter
	recycleAdapterOnce sync.Once
)

// NewRecycleSystemAdapter 创建回收系统适配器
func NewRecycleSystemAdapter() *RecycleSystemAdapter {
	recycleUC := recycleusecase.NewRecycleItemUseCase(deps.ConfigGateway())
	recycleUC.SetDependencies(NewBagUseCaseAdapter(), reward.NewRewardUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway()))
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

// GetRecycleSys 获取回收系统适配器（保持接口一致性）
func GetRecycleSys(ctx context.Context) *RecycleSystemAdapter {
	_ = ctx // 当前回收系统无状态，仅保留参数以保持接口一致
	return getRecycleSysInstance()
}

func init() {
	// 协议注册由 controller 包负责
}
