package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// SysMgr 系统管理器
type SysMgr struct {
	factories map[uint32]iface.SystemFactory // 系统工厂
	sysList   []iface.ISystem                // 系统列表（按系统ID索引）
}

var (
	globalFactories = make(map[uint32]iface.SystemFactory)
)

// RegisterSystemFactory 注册系统工厂（全局注册）
func RegisterSystemFactory(sysId uint32, factory iface.SystemFactory) {
	globalFactories[sysId] = factory
}

// NewSysMgr 创建系统管理器
func NewSysMgr() iface.ISystemMgr {
	mgr := &SysMgr{
		sysList:   make([]iface.ISystem, protocol.SystemId_SysIdMax),
		factories: make(map[uint32]iface.SystemFactory),
	}
	// 复制全局工厂
	for sysId, factory := range globalFactories {
		mgr.factories[sysId] = factory
	}
	return mgr
}

func (m *SysMgr) OnInit(ctx context.Context) error {
	for sysId := uint32(1); sysId < uint32(protocol.SystemId_SysIdMax); sysId++ {
		factory := m.factories[sysId]
		if factory == nil {
			log.Errorf("sys:%d not found system factory", sysId)
			continue
		}
		system := factory()
		system.OnInit(ctx)
		system.SetOpened(true)
		m.sysList[sysId] = system
	}
	return nil
}

func (m *SysMgr) OnRoleLogin(ctx context.Context) {
	m.CheckAllSysOpen(ctx)
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnRoleLogin(ctx)
	})
}

func (m *SysMgr) OnRoleReconnect(ctx context.Context) {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnRoleReconnect(ctx)
	})
}

// GetSystem 获取系统
func (m *SysMgr) GetSystem(sysId uint32) iface.ISystem {
	if sysId <= 0 || sysId >= uint32(protocol.SystemId_SysIdMax) {
		return nil
	}
	return m.sysList[sysId]
}

func (m *SysMgr) CheckAllSysOpen(ctx context.Context) {
	iPlayerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}
	for _, system := range m.sysList {
		if system == nil {
			continue
		}
		if !system.IsOpened() {
			continue
		}
		iPlayerRole.SetSysStatus(system.GetId(), true)
		system.SetOpened(true)
	}
	return
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
		// 串行执行，不创建新协程，保持单Actor模型
		f(system)
	}
}

func handleSysMgrOnPlayerLogin(ctx context.Context, _ *event.Event) {
	iPlayerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}
	mgr := iPlayerRole.GetSysMgr().(*SysMgr)
	mgr.OnRoleLogin(ctx)
}

func handleSysMgrOnRoleReconnect(ctx context.Context, _ *event.Event) {
	iPlayerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}
	mgr := iPlayerRole.GetSysMgr().(*SysMgr)
	mgr.OnRoleReconnect(ctx)
}

func init() {
	gevent.SubscribePlayerEvent(gevent.OnPlayerLogin, handleSysMgrOnPlayerLogin)
	gevent.SubscribePlayerEvent(gevent.OnPlayerReconnect, handleSysMgrOnRoleReconnect)
}
