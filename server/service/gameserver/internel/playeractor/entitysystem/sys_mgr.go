package entitysystem

import (
	"fmt"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"postapocgame/server/service/gameserver/internel/iface"
)

// SysMgr 系统管理器
type SysMgr struct {
	role          iface.IPlayerRole
	sysList       []iface.ISystem                // 系统列表（按系统ID索引）
	sysOpenStatus map[uint32]bool                // 系统开启状态
	factories     map[uint32]iface.SystemFactory // 系统工厂
}

var (
	globalFactories = make(map[uint32]iface.SystemFactory)
)

// RegisterSystemFactory 注册系统工厂（全局注册）
func RegisterSystemFactory(sysID uint32, factory iface.SystemFactory) {
	globalFactories[sysID] = factory
}

// NewSysMgr 创建系统管理器
func NewSysMgr(role iface.IPlayerRole) *SysMgr {
	mgr := &SysMgr{
		role:          role,
		sysList:       make([]iface.ISystem, protocol.SystemId_SysIdMax),
		sysOpenStatus: make(map[uint32]bool),
		factories:     make(map[uint32]iface.SystemFactory),
	}

	// 复制全局工厂
	for sysID, factory := range globalFactories {
		mgr.factories[sysID] = factory
	}

	// 按系统ID从小到大的顺序初始化
	for sysID := uint32(1); sysID < uint32(protocol.SystemId_SysIdMax); sysID++ {
		factory, ok := mgr.factories[sysID]
		if !ok {
			log.Warnf("System factory not found: sysID=%d", sysID)
			continue
		}

		system := factory()
		mgr.sysList[sysID] = system
		mgr.sysOpenStatus[sysID] = true // 默认都开启
	}

	return mgr
}

// onRoleLogin 角色登录事件处理
func (m *SysMgr) onRoleLogin() error {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnOpen()
	})

	return nil
}

// onRoleLogout 角色登出事件处理
func (m *SysMgr) onRoleLogout() error {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnRoleLogout()
	})
	return nil
}

// OnReconnect 重连时调用所有系统的 OnRoleReconnect
func (m *SysMgr) OnReconnect() error {
	for _, system := range m.sysList {
		if system == nil {
			continue
		}

		if !m.IsSystemOpened(system.GetID()) {
			continue
		}

		system.OnRoleReconnect()
	}

	return nil
}

// OnClose 关闭时调用所有系统的 OnRoleClose
func (m *SysMgr) OnClose() error {
	for _, system := range m.sysList {
		if system == nil {
			continue
		}

		system.OnRoleClose()
	}

	return nil
}

// GetSystem 获取系统
func (m *SysMgr) GetSystem(sysID uint32) iface.ISystem {
	if sysID <= 0 || sysID >= uint32(protocol.SystemId_SysIdMax) {
		return nil
	}
	return m.sysList[sysID]
}

// OpenSystem 开启系统
func (m *SysMgr) OpenSystem(sysID uint32) error {
	system := m.GetSystem(sysID)
	if system == nil {
		return fmt.Errorf("system not found: sysID=%d", sysID)
	}

	if m.IsSystemOpened(sysID) {
		return nil // 已经开启
	}

	system.OnOpen()

	m.sysOpenStatus[sysID] = true
	log.Infof("System opened: sysID=%d", sysID)
	return nil
}

// CloseSystem 关闭系统
func (m *SysMgr) CloseSystem(sysID uint32) {
	m.sysOpenStatus[sysID] = false
	log.Infof("System closed: sysID=%d", sysID)
}

// IsSystemOpened 检查系统是否开启
func (m *SysMgr) IsSystemOpened(sysID uint32) bool {
	opened, ok := m.sysOpenStatus[sysID]
	if !ok {
		return true // 默认开启
	}
	return opened
}

// GetOpenedSystems 获取所有已开启的系统
func (m *SysMgr) GetOpenedSystems() []iface.ISystem {
	var systems []iface.ISystem
	for _, system := range m.sysList {
		if system == nil {
			continue
		}
		if m.IsSystemOpened(system.GetID()) {
			systems = append(systems, system)
		}
	}
	return systems
}

func (m *SysMgr) EachOpenSystem(f func(system iface.ISystem)) {
	if f == nil {
		return
	}
	for _, system := range m.sysList {
		if system == nil {
			continue
		}
		if !system.IsOpened() {
			continue
		}
		routine.Run(func() {
			f(system)
		})
	}
}
