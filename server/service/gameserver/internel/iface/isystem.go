package iface

import (
	"postapocgame/server/internal/custom_id"
)

// ISystem 系统接口
type ISystem interface {
	GetID() custom_id.SystemId
	OnOpen()
	OnRoleLogin()
	OnRoleReconnect()
	OnRoleLogout()
	OnRoleClose()
	IsOpened() bool
	SetOpened(opened bool)
}

// SystemFactory 系统工厂函数
type SystemFactory func(role IPlayerRole) ISystem
