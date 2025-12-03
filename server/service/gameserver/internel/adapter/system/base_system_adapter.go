package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/iface"
)

// BaseSystemAdapter 系统适配器基类
//
// SystemAdapter 职责说明（Clean Architecture 原则）：
// - 生命周期适配：将 Actor 生命周期事件（OnInit/RunOne/OnNewDay 等）转换为对 UseCase 的调用
// - 事件订阅：订阅玩家级事件，在事件到来时调用相应的 UseCase
// - 状态管理：管理与 Actor 运行模型强相关的运行时状态（如 dirty 标记、定时任务）
// - 禁止：直接操作数据库、网络，或实现业务规则逻辑（业务逻辑应在 UseCase 层）
//
// 生命周期方法说明：
// - OnInit: 系统初始化时调用，通常用于调用 InitDataUseCase 初始化数据结构
// - OnRoleLogin: 玩家登录时调用，用于加载/推送数据
// - OnRoleReconnect: 玩家重连时调用，用于同步数据
// - OnRoleLogout: 玩家登出时调用，用于保存/清理数据
// - OnRoleClose: 玩家关闭时调用，用于最终清理
// - RunOne: 每帧调用，用于定期检查、定时任务、属性同步等
// - OnNewDay/OnNewWeek/OnNewMonth: 时间事件，用于刷新每日/每周/每月数据
//
// 注意：所有生命周期方法应只做"何时调用哪个 UseCase"的调度，具体业务逻辑由 UseCase 承担。
//
// ⚠️ 防退化机制（重要）：
// - 禁止在 SystemAdapter 中编写业务规则逻辑（如校验、计算、状态转换等）
// - 禁止直接操作数据库、网络或配置（应通过 Repository/Gateway 接口）
// - 所有业务逻辑必须在 UseCase 层实现，SystemAdapter 只负责"何时调用哪个 UseCase"
// - 如果发现 SystemAdapter 中有业务逻辑，应立即重构到 UseCase 层
// - Code Review 时必须检查 SystemAdapter 是否含有可下沉到 UseCase 的逻辑
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
