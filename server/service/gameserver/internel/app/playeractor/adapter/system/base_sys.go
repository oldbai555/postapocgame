package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
)

type BaseSystemAdapter struct {
	sysID  uint32
	opened bool
}

// NewBaseSystemAdapter 创建系统适配器基类
func NewBaseSystemAdapter(sysID uint32) *BaseSystemAdapter {
	return &BaseSystemAdapter{
		sysID:  sysID,
		opened: true, // 默认开启
	}
}

// GetId 获取系统ID
func (a *BaseSystemAdapter) GetId() uint32 {
	return a.sysID
}

// IsOpened 获取系统开启状态
func (a *BaseSystemAdapter) IsOpened() bool {
	return a.opened
}

// SetOpened 设置系统开启状态
// 注意：此方法在系统初始化时调用，此时可能还没有 Context
// 系统状态会在 CheckAllSysOpen 时统一更新到 BinaryData
func (a *BaseSystemAdapter) SetOpened(opened bool) {
	a.opened = opened
	// 注意：系统状态的持久化由 SysMgr.CheckAllSysOpen 统一处理
	// 这里只更新内存状态
}

// OnInit 系统初始化（子类可以重写）
func (a *BaseSystemAdapter) OnInit(ctx context.Context) {}

// OnOpen 系统开启（子类可以重写）
func (a *BaseSystemAdapter) OnOpen(ctx context.Context) {}

// OnRoleLogin 玩家登录（子类可以重写）
func (a *BaseSystemAdapter) OnRoleLogin(ctx context.Context) {}

// OnRoleReconnect 玩家重连（子类可以重写）
func (a *BaseSystemAdapter) OnRoleReconnect(ctx context.Context) {}

// OnRoleLogout 玩家登出（子类可以重写）
func (a *BaseSystemAdapter) OnRoleLogout(ctx context.Context) {}

// OnRoleClose 玩家关闭（子类可以重写）
func (a *BaseSystemAdapter) OnRoleClose(ctx context.Context) {}

// OnNewHour 新小时（子类可以重写）
func (a *BaseSystemAdapter) OnNewHour(ctx context.Context) {}

// OnNewDay 新天（子类可以重写）
func (a *BaseSystemAdapter) OnNewDay(ctx context.Context) {}

// OnNewWeek 新周（子类可以重写）
func (a *BaseSystemAdapter) OnNewWeek(ctx context.Context) {}

// OnNewMonth 新月（子类可以重写）
func (a *BaseSystemAdapter) OnNewMonth(ctx context.Context) {}

// OnNewYear 新年（子类可以重写）
func (a *BaseSystemAdapter) OnNewYear(ctx context.Context) {}

// RunOne 每帧调用（子类可以重写）
func (a *BaseSystemAdapter) RunOne(ctx context.Context) {}

// EnsureISystem 确保 BaseSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*BaseSystemAdapter)(nil)
