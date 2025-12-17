package recycle

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/app/playeractor/sysbase"
	"sync"
)

// RecycleSystemAdapter 回收系统适配器（单例模式）
type RecycleSystemAdapter struct {
	*sysbase.BaseSystem
	deps               Deps
	recycleItemUseCase *RecycleItemUseCase
}

var (
	recycleAdapter     *RecycleSystemAdapter
	recycleAdapterOnce sync.Once
	recycleRuntime     *runtime.Runtime // 保存 Runtime 实例用于单例初始化
)

// NewRecycleSystemAdapter 创建回收系统适配器
func NewRecycleSystemAdapter(rt *runtime.Runtime) *RecycleSystemAdapter {
	d := depsFromRuntime(rt)
	return &RecycleSystemAdapter{
		// Recycle 目前并未纳入 SystemId 枚举，这里使用 0 作为占位 ID，
		// 仅用于满足 ISystem 接口要求，不参与系统开关与持久化。
		BaseSystem:         sysbase.NewBaseSystem(uint32(protocol.SystemId_SystemIdNil)),
		deps:               d,
		recycleItemUseCase: NewRecycleItemUseCase(d),
	}
}

// getRecycleSysInstance 获取回收系统适配器实例
func getRecycleSysInstance() *RecycleSystemAdapter {
	recycleAdapterOnce.Do(func() {
		if recycleRuntime != nil {
			recycleAdapter = NewRecycleSystemAdapter(recycleRuntime)
		}
	})
	return recycleAdapter
}

// RegisterSystemFactory 注册回收系统工厂（由 register.RegisterAll 调用）
func RegisterSystemFactory(rt *runtime.Runtime) {
	// Recycle 系统使用单例模式，需要保存 Runtime 实例
	recycleRuntime = rt
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

// init() 函数已移除，注册逻辑迁移至 playeractor/register.RegisterAll()
