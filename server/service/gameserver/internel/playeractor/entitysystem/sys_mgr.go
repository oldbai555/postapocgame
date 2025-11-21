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
	// 系统依赖关系：key为系统ID，value为依赖的系统ID列表
	// 例如：AttrSys依赖LevelSys和EquipSys
	systemDependencies = map[uint32][]uint32{
		uint32(protocol.SystemId_SysAttr): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysEquip),
		},
		// 可以在这里添加更多依赖关系
	}
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
	// 使用拓扑排序确定系统初始化顺序
	initOrder := m.getInitOrder()

	// 按照依赖顺序初始化系统
	for _, sysId := range initOrder {
		factory := m.factories[sysId]
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

// getInitOrder 使用拓扑排序获取系统初始化顺序
func (m *SysMgr) getInitOrder() []uint32 {
	// 构建依赖图
	inDegree := make(map[uint32]int)   // 入度表
	graph := make(map[uint32][]uint32) // 依赖图：key依赖value列表中的系统

	// 初始化所有系统的入度为0
	for sysId := uint32(1); sysId < uint32(protocol.SystemId_SysIdMax); sysId++ {
		if m.factories[sysId] != nil {
			inDegree[sysId] = 0
		}
	}

	// 构建依赖图
	for sysId, deps := range systemDependencies {
		if m.factories[sysId] == nil {
			continue
		}
		for _, depId := range deps {
			if m.factories[depId] != nil {
				graph[depId] = append(graph[depId], sysId)
				inDegree[sysId]++
			}
		}
	}

	// 拓扑排序
	var queue []uint32
	var result []uint32

	// 将所有入度为0的系统加入队列
	for sysId, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, sysId)
		}
	}

	// BFS拓扑排序
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 处理依赖当前系统的系统
		for _, nextSysId := range graph[current] {
			inDegree[nextSysId]--
			if inDegree[nextSysId] == 0 {
				queue = append(queue, nextSysId)
			}
		}
	}

	// 如果还有系统未处理，说明存在循环依赖（不应该发生）
	if len(result) < len(inDegree) {
		log.Errorf("System dependency cycle detected! Initialized: %d, Total: %d", len(result), len(inDegree))
		// 将未处理的系统按原顺序加入（作为fallback）
		for sysId := uint32(1); sysId < uint32(protocol.SystemId_SysIdMax); sysId++ {
			if m.factories[sysId] != nil {
				found := false
				for _, r := range result {
					if r == sysId {
						found = true
						break
					}
				}
				if !found {
					result = append(result, sysId)
				}
			}
		}
	}

	log.Infof("System init order: %v", result)
	return result
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
		log.Errorf("get player role error:%v", err)
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
