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
	// 系统依赖关系：key为系统ID，value为依赖的系统ID列表
	// 说明：
	// - 仅在此处维护「系统之间」的依赖关系，用于确定初始化顺序
	// - UseCase 之间的依赖（MoneyUseCase / RewardUseCase 等）通过 DI 容器注入，不在此表中体现
	systemDependencies = map[uint32][]uint32{
		// 等级系统：无前置依赖
		// SystemId_SysLevel: {},

		// 背包系统：依赖等级系统（需要等级信息做容量/功能开放判断）
		uint32(protocol.SystemId_SysBag): {
			uint32(protocol.SystemId_SysLevel),
		},

		// 货币系统：依赖等级系统（部分货币/解锁逻辑与等级相关）
		uint32(protocol.SystemId_SysMoney): {
			uint32(protocol.SystemId_SysLevel),
		},

		// GM 系统：依赖背包/货币/等级，用于发放奖励与修改属性
		uint32(protocol.SystemId_SysGM): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysBag),
			uint32(protocol.SystemId_SysMoney),
		},

		// 技能系统：依赖等级系统（学习/升级技能的等级条件）
		uint32(protocol.SystemId_SysSkill): {
			uint32(protocol.SystemId_SysLevel),
		},

		// 属性系统：依赖等级与装备系统（属性来源）
		uint32(protocol.SystemId_SysAttr): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysEquip),
		},

		// 装备系统：依赖背包系统（穿戴/脱装备需要操作背包物品）
		uint32(protocol.SystemId_SysEquip): {
			uint32(protocol.SystemId_SysBag),
		},

		// VIP 系统：依赖货币系统（VIP 经验通过特殊货币累积）
		uint32(protocol.SystemId_SysVip): {
			uint32(protocol.SystemId_SysMoney),
		},

		// 任务系统：依赖等级/属性/背包/货币/日常活跃
		uint32(protocol.SystemId_SysQuest): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysAttr),
			uint32(protocol.SystemId_SysBag),
			uint32(protocol.SystemId_SysMoney),
			uint32(protocol.SystemId_SysDailyActivity),
		},

		// 邮件系统：依赖背包/货币系统（领取附件发放物品与货币）
		uint32(protocol.SystemId_SysMail): {
			uint32(protocol.SystemId_SysBag),
			uint32(protocol.SystemId_SysMoney),
		},

		// 商城系统：依赖背包/货币系统（购买消耗货币发放物品）
		uint32(protocol.SystemId_SysShop): {
			uint32(protocol.SystemId_SysBag),
			uint32(protocol.SystemId_SysMoney),
		},

		// 副本系统：依赖等级/属性/背包/货币系统
		uint32(protocol.SystemId_SysFuBen): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysAttr),
			uint32(protocol.SystemId_SysBag),
			uint32(protocol.SystemId_SysMoney),
		},

		// 物品使用系统：依赖背包/等级/属性/货币系统
		uint32(protocol.SystemId_SysItemUse): {
			uint32(protocol.SystemId_SysBag),
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysAttr),
			uint32(protocol.SystemId_SysMoney),
		},

		// 日常活跃系统：依赖等级系统（活跃任务解锁）
		uint32(protocol.SystemId_SysDailyActivity): {
			uint32(protocol.SystemId_SysLevel),
		},

		// 好友系统：依赖等级系统（部分好友功能可能存在等级门槛）
		uint32(protocol.SystemId_SysFriend): {
			uint32(protocol.SystemId_SysLevel),
		},

		// 公会系统：依赖等级系统（创建/加入条件）与好友系统（邀请/推荐等）
		uint32(protocol.SystemId_SysGuild): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysFriend),
		},

		// 拍卖行系统：依赖背包/货币系统
		uint32(protocol.SystemId_SysAuction): {
			uint32(protocol.SystemId_SysBag),
			uint32(protocol.SystemId_SysMoney),
		},

		// 聊天系统：依赖等级系统（世界频道解锁）与好友/公会系统
		uint32(protocol.SystemId_SysChat): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysFriend),
			uint32(protocol.SystemId_SysGuild),
		},

		// 排行榜系统：依赖等级/属性/公会等数据
		uint32(protocol.SystemId_SysRank): {
			uint32(protocol.SystemId_SysLevel),
			uint32(protocol.SystemId_SysAttr),
			uint32(protocol.SystemId_SysGuild),
		},

		// 玩家消息系统：依赖邮件/任务/副本等系统，用于回放与通知
		uint32(protocol.SystemId_SysMessage): {
			uint32(protocol.SystemId_SysMail),
			uint32(protocol.SystemId_SysQuest),
			uint32(protocol.SystemId_SysFuBen),
		},
	}
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
