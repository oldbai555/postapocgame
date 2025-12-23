package entitysystem

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/iface"
	"sync"
)

// SystemRegistry 系统注册表（替代全局 globalFactories）
type SystemRegistry struct {
	factories map[uint32]iface.SystemFactory
	mu        sync.RWMutex
}

var (
	globalRegistry = &SystemRegistry{
		factories: make(map[uint32]iface.SystemFactory),
	}
)

// RegisterSystemFactory 注册系统工厂（模块级注册，替代全局注册）
func RegisterSystemFactory(sysId uint32, factory iface.SystemFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[sysId] = factory
}

// GetSystemFactory 获取系统工厂
func (r *SystemRegistry) GetSystemFactory(sysId uint32) iface.SystemFactory {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.factories[sysId]
}

// GetAllSystemIds 获取所有已注册的系统ID列表（用于显式配置）
func (r *SystemRegistry) GetAllSystemIds() []uint32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]uint32, 0, len(r.factories))
	for id := range r.factories {
		ids = append(ids, id)
	}
	return ids
}

// GetDefaultSystemIds 获取默认需要挂载的系统ID列表（按依赖顺序）
// 这个列表可以在配置文件中配置，或者通过环境变量控制
func GetDefaultSystemIds() []uint32 {
	return []uint32{
		uint32(protocol.SystemId_SysLevel),
		uint32(protocol.SystemId_SysSkill),
	}
}
