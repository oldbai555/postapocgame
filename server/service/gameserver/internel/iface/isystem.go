package iface

// ISystem 系统接口
type ISystem interface {
	GetId() uint32
	OnOpen()
	OnRoleLogin()
	OnRoleReconnect()
	OnRoleLogout()
	OnRoleClose()
	IsOpened() bool
	SetOpened(opened bool)
}

// SystemFactory 系统工厂函数
type SystemFactory func() ISystem
