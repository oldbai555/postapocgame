package entitysystem

// BaseSystem 系统基类
type BaseSystem struct {
	sysID  uint32
	opened bool
}

func NewBaseSystem(sysID uint32) *BaseSystem {
	return &BaseSystem{
		sysID:  sysID,
		opened: true, // 默认开启
	}
}

func (bs *BaseSystem) GetId() uint32 {
	return bs.sysID
}

func (bs *BaseSystem) OnOpen() {
	bs.opened = true
	return
}

func (bs *BaseSystem) OnRoleLogin() {
	// 基类默认实现：什么都不做
	return
}

func (bs *BaseSystem) OnRoleReconnect() {
	return
}

func (bs *BaseSystem) OnRoleLogout() {
	return
}

func (bs *BaseSystem) OnRoleClose() {
	return
}

func (bs *BaseSystem) IsOpened() bool {
	return bs.opened
}

func (bs *BaseSystem) SetOpened(opened bool) {
	bs.opened = opened
}
