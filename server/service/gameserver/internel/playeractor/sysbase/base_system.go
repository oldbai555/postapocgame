package sysbase

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
)

// BaseSystem 是 PlayerActor 下各业务系统共享的基础实现。
// 它完整实现了 iface.ISystem 接口，子系统只需要在需要的生命周期钩子上做覆盖即可。
type BaseSystem struct {
	sysID  uint32
	opened bool
}

// NewBaseSystem 创建一个默认开启的系统基类
func NewBaseSystem(sysID uint32) *BaseSystem {
	return &BaseSystem{
		sysID:  sysID,
		opened: true, // 默认开启
	}
}

// GetId 获取系统 ID
func (s *BaseSystem) GetId() uint32 {
	return s.sysID
}

// IsOpened 获取系统开启状态
func (s *BaseSystem) IsOpened() bool {
	return s.opened
}

// SetOpened 设置系统开启状态
// 注意：此方法在系统初始化/开关检查时调用，持久化由 SysMgr 统一处理。
func (s *BaseSystem) SetOpened(opened bool) {
	s.opened = opened
}

// 以下生命周期方法均为空实现，便于子系统按需覆盖。

func (s *BaseSystem) OnInit(ctx context.Context)       {}
func (s *BaseSystem) OnOpen(ctx context.Context)       {}
func (s *BaseSystem) OnRoleLogin(ctx context.Context)  {}
func (s *BaseSystem) OnRoleLogout(ctx context.Context) {}
func (s *BaseSystem) OnRoleClose(ctx context.Context)  {}
func (s *BaseSystem) OnNewHour(ctx context.Context)    {}
func (s *BaseSystem) OnNewDay(ctx context.Context)     {}
func (s *BaseSystem) OnNewWeek(ctx context.Context)    {}
func (s *BaseSystem) OnNewMonth(ctx context.Context)   {}
func (s *BaseSystem) OnNewYear(ctx context.Context)    {}

// 确保 BaseSystem 实现 ISystem 接口
var _ iface.ISystem = (*BaseSystem)(nil)
