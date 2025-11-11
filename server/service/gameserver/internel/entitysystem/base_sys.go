package entitysystem

import (
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/service/gameserver/internel/iface"
)

// BaseSystem 系统基类
type BaseSystem struct {
	sysID  custom_id.SystemId
	role   iface.IPlayerRole
	opened bool
}

// NewBaseSystem 创建基础系统
func NewBaseSystem(sysID custom_id.SystemId, role iface.IPlayerRole) *BaseSystem {
	return &BaseSystem{
		sysID:  sysID,
		role:   role,
		opened: true, // 默认开启
	}
}

// GetID 获取系统ID
func (bs *BaseSystem) GetID() custom_id.SystemId {
	return bs.sysID
}

// OnOpen 首次开启系统（子类可重写）
func (bs *BaseSystem) OnOpen() {
	bs.opened = true
	return
}

// OnRoleLogin 角色登录（子类必须重写）
func (bs *BaseSystem) OnRoleLogin() {
	// 基类默认实现：什么都不做
	return
}

// OnRoleReconnect 角色重连（子类可重写，默认和登录相同）
func (bs *BaseSystem) OnRoleReconnect() {
	return
}

// OnRoleLogout 角色登出（子类可重写）
func (bs *BaseSystem) OnRoleLogout() {
	return
}

// OnRoleClose 角色关闭（子类可重写）
func (bs *BaseSystem) OnRoleClose() {
	return
}

// IsOpened 是否已开启
func (bs *BaseSystem) IsOpened() bool {
	return bs.opened
}

// SetOpened 设置开启状态
func (bs *BaseSystem) SetOpened(opened bool) {
	bs.opened = opened
}

// GetRole 获取角色引用
func (bs *BaseSystem) GetRole() iface.IPlayerRole {
	return bs.role
}
