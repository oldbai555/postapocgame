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
func RegisterSystemFactory(sysId uint32, factory iface.SystemFactory) {
	globalFactories[sysId] = factory
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
	for sysId, factory := range globalFactories {
		mgr.factories[sysId] = factory
	}

	// 按系统ID从小到大的顺序初始化
	for sysId := uint32(1); sysId < uint32(protocol.SystemId_SysIdMax); sysId++ {
		factory, ok := mgr.factories[sysId]
		if !ok {
			log.Warnf("System factory not found: sysId=%d", sysId)
			continue
		}

		system := factory()
		mgr.sysList[sysId] = system
		mgr.sysOpenStatus[sysId] = true // 默认都开启
	}

	return mgr
}

// GetSystem 获取系统
func (m *SysMgr) GetSystem(sysId uint32) iface.ISystem {
	if sysId <= 0 || sysId >= uint32(protocol.SystemId_SysIdMax) {
		return nil
	}
	return m.sysList[sysId]
}

// OpenSystem 开启系统
func (m *SysMgr) OpenSystem(sysId uint32) error {
	system := m.GetSystem(sysId)
	if system == nil {
		return fmt.Errorf("system not found: sysId=%d", sysId)
	}

	if m.IsSystemOpened(sysId) {
		return nil // 已经开启
	}

	system.OnOpen()

	m.sysOpenStatus[sysId] = true
	log.Infof("System opened: sysId=%d", sysId)
	return nil
}

// IsSystemOpened 检查系统是否开启
func (m *SysMgr) IsSystemOpened(sysId uint32) bool {
	opened, ok := m.sysOpenStatus[sysId]
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
		if m.IsSystemOpened(system.GetId()) {
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
