package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"

	"postapocgame/server/service/gameserver/internel/gevent"
)

// SysMgr 系统管理器
type SysMgr struct {
	systemIds []uint32        // 显式配置的系统ID列表
	sysList   []iface.ISystem // 系统列表（按系统ID索引）
}

// NewSysMgr 创建系统管理器（使用默认系统列表）
func NewSysMgr() iface.ISystemMgr {
	return NewSysMgrWithSystems(GetDefaultSystemIds())
}

// NewSysMgrWithSystems 创建系统管理器（显式指定系统列表）
func NewSysMgrWithSystems(systemIds []uint32) iface.ISystemMgr {
	mgr := &SysMgr{
		systemIds: systemIds,
		sysList:   make([]iface.ISystem, protocol.SystemId_SysIdMax),
	}
	return mgr
}

func (m *SysMgr) OnInit(ctx context.Context) error {
	// 只初始化显式配置的系统，不再遍历所有枚举值
	for _, sysId := range m.systemIds {
		factory := globalRegistry.GetSystemFactory(sysId)
		if factory == nil {
			log.Warnf("System factory not found for SysId=%d, skipping", sysId)
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
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnRoleLogin(ctx)
	})
}

func (m *SysMgr) OnRoleReconnect(ctx context.Context) {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnRoleReconnect(ctx)
	})
}

func (m *SysMgr) OnNewHour(ctx context.Context) {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnNewHour(ctx)
	})
}

func (m *SysMgr) OnNewDay(ctx context.Context) {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnNewDay(ctx)
	})
}

func (m *SysMgr) OnNewWeek(ctx context.Context) {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnNewWeek(ctx)
	})
}

func (m *SysMgr) OnNewMonth(ctx context.Context) {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnNewMonth(ctx)
	})
}

func (m *SysMgr) OnNewYear(ctx context.Context) {
	m.EachOpenSystem(func(system iface.ISystem) {
		system.OnNewYear(ctx)
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

// ListMountedSystems 列出已挂载的系统（调试函数）
func (m *SysMgr) ListMountedSystems() []SystemInfo {
	infos := make([]SystemInfo, 0)
	for _, sysId := range m.systemIds {
		system := m.sysList[sysId]
		if system == nil {
			continue
		}
		infos = append(infos, SystemInfo{
			SysId:   sysId,
			Opened:  system.IsOpened(),
			HasImpl: true,
		})
	}
	return infos
}

// SystemInfo 系统信息（用于调试）
type SystemInfo struct {
	SysId   uint32
	Opened  bool
	HasImpl bool
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
	gevent.SubscribePlayerEvent(gevent.OnPlayerLogin, handleSysMgrOnPlayerLogin)
	gevent.SubscribePlayerEvent(gevent.OnPlayerReconnect, handleSysMgrOnRoleReconnect)
}

// GetIPlayerRoleByContext 从上下文中解析玩家角色（兼容旧 EntitySystem 代码）
func GetIPlayerRoleByContext(ctx context.Context) (iface.IPlayerRole, error) {
	if ctx == nil {
		return nil, customerr.NewError("context is nil")
	}
	val := ctx.Value(gshare.ContextKeyRole)
	if val == nil {
		return nil, customerr.NewError("no player role in context")
	}
	playerRole, ok := val.(iface.IPlayerRole)
	if !ok {
		return nil, customerr.NewError("context value is not iface.IPlayerRole")
	}
	return playerRole, nil
}
