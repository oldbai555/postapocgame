package entitysystem

import (
	"context"
	"fmt"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	iface2 "postapocgame/server/service/gameserver/internel/core/iface"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"

	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
)

// SysMgr 系统管理器
type SysMgr struct {
	factories map[uint32]iface2.SystemFactory // 系统工厂
	sysList   []iface2.ISystem                // 系统列表（按系统ID索引）
}

var (
	globalFactories = make(map[uint32]iface2.SystemFactory)
)

// RegisterSystemFactory 注册系统工厂（全局注册）
func RegisterSystemFactory(sysId uint32, factory iface2.SystemFactory) {
	globalFactories[sysId] = factory
}

// NewSysMgr 创建系统管理器
func NewSysMgr() iface2.ISystemMgr {
	mgr := &SysMgr{
		sysList:   make([]iface2.ISystem, protocol.SystemId_SysIdMax),
		factories: make(map[uint32]iface2.SystemFactory),
	}
	// 复制全局工厂
	for sysId, factory := range globalFactories {
		mgr.factories[sysId] = factory
	}
	return mgr
}

func (m *SysMgr) OnInit(ctx context.Context) error {
	// 按照依赖顺序初始化系统
	for sysId := protocol.SystemId_SysLevel; sysId < protocol.SystemId_SysIdMax; sysId++ {
		factory := m.factories[uint32(sysId)]
		if factory == nil {
			log.Errorf("sys:%d not found system factory", sysId)
			continue
		}
		system := factory()
		system.OnInit(ctx)
		system.SetOpened(true)
		m.sysList[sysId] = system
		log.Debugf("System initialized: SysId=%d", sysId)
	}
	return nil
}

func (m *SysMgr) OnRoleLogin(ctx context.Context) {
	m.CheckAllSysOpen(ctx)
	m.EachOpenSystem(func(system iface2.ISystem) {
		system.OnRoleLogin(ctx)
	})
}

func (m *SysMgr) OnRoleReconnect(ctx context.Context) {
	m.EachOpenSystem(func(system iface2.ISystem) {
		system.OnRoleReconnect(ctx)
	})
}

func (m *SysMgr) OnNewHour(ctx context.Context) {
	m.EachOpenSystem(func(system iface2.ISystem) {
		system.OnNewHour(ctx)
	})
}

func (m *SysMgr) OnNewDay(ctx context.Context) {
	m.EachOpenSystem(func(system iface2.ISystem) {
		system.OnNewDay(ctx)
	})
}

func (m *SysMgr) OnNewWeek(ctx context.Context) {
	m.EachOpenSystem(func(system iface2.ISystem) {
		system.OnNewWeek(ctx)
	})
}

func (m *SysMgr) OnNewMonth(ctx context.Context) {
	m.EachOpenSystem(func(system iface2.ISystem) {
		system.OnNewMonth(ctx)
	})
}

func (m *SysMgr) OnNewYear(ctx context.Context) {
	m.EachOpenSystem(func(system iface2.ISystem) {
		system.OnNewYear(ctx)
	})
}

// GetSystem 获取系统
func (m *SysMgr) GetSystem(sysId uint32) iface2.ISystem {
	if sysId <= 0 || sysId >= uint32(protocol.SystemId_SysIdMax) {
		return nil
	}
	return m.sysList[sysId]
}

func (m *SysMgr) CheckAllSysOpen(ctx context.Context) {
	iPlayerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("CheckAllSysOpen: get player role error: %v", err)
		return
	}
	for _, system := range m.sysList {
		if system == nil {
			continue
		}
		if system.IsOpened() {
			continue
		}
		iPlayerRole.SetSysStatus(system.GetId(), true)
		system.SetOpened(true)
	}
	return
}

func (m *SysMgr) EachOpenSystem(f func(system iface2.ISystem)) {
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
		log.Errorf("handleSysMgrOnPlayerLogin: get player role error:%v", err)
		return
	}
	mgr := iPlayerRole.GetSysMgr().(*SysMgr)
	mgr.OnRoleLogin(ctx)
}

func handleSysMgrOnRoleReconnect(ctx context.Context, _ *event.Event) {
	iPlayerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleSysMgrOnRoleReconnect: get player role error:%v", err)
		return
	}
	mgr := iPlayerRole.GetSysMgr().(*SysMgr)
	mgr.OnRoleReconnect(ctx)
}

func init() {
	gevent2.SubscribePlayerEvent(gevent2.OnPlayerLogin, handleSysMgrOnPlayerLogin)
	gevent2.SubscribePlayerEvent(gevent2.OnPlayerReconnect, handleSysMgrOnRoleReconnect)
}

// GetIPlayerRoleByContext 从上下文中解析玩家角色（兼容旧 EntitySystem 代码）
func GetIPlayerRoleByContext(ctx context.Context) (iface2.IPlayerRole, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	val := ctx.Value(gshare.ContextKeyRole)
	if val == nil {
		return nil, fmt.Errorf("no player role in context")
	}
	playerRole, ok := val.(iface2.IPlayerRole)
	if !ok {
		return nil, fmt.Errorf("context value is not iface.IPlayerRole")
	}
	return playerRole, nil
}
