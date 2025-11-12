package iface

import "context"

// ISystem 系统接口
type ISystem interface {
	GetId() uint32
	OnInit(ctx context.Context)
	OnOpen(ctx context.Context)
	OnRoleLogin(ctx context.Context)
	OnRoleReconnect(ctx context.Context)
	OnRoleLogout(ctx context.Context)
	OnRoleClose(ctx context.Context)
	IsOpened() bool
	SetOpened(opened bool)
}

type ISystemMgr interface {
	GetSystem(sysId uint32) ISystem

	OnInit(ctx context.Context) error
	CheckAllSysOpen(ctx context.Context)
	OnRoleLogin(ctx context.Context)
	OnRoleReconnect(ctx context.Context)
}

// SystemFactory 系统工厂函数
type SystemFactory func() ISystem
