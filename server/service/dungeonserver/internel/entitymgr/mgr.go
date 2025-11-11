package entitymgr

import (
	"fmt"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
	"sync/atomic"
)

// EntityMgr 全局实体管理器
type EntityMgr struct {
	mu       sync.RWMutex
	entities map[uint64]iface.IEntity // hdl -> entity

	nextHdl uint64 // 下一个句柄
}

var (
	globalEntityMgr *EntityMgr
	entityMgrOnce   sync.Once
)

// GetEntityMgr 获取全局实体管理器
func GetEntityMgr() *EntityMgr {
	entityMgrOnce.Do(func() {
		globalEntityMgr = &EntityMgr{
			entities: make(map[uint64]iface.IEntity),
			nextHdl:  100000, // 从100000开始
		}
	})
	return globalEntityMgr
}

// GenHdl 生成唯一句柄
func (m *EntityMgr) GenHdl() uint64 {
	return atomic.AddUint64(&m.nextHdl, 1)
}

// Register 注册实体
func (m *EntityMgr) Register(entity iface.IEntity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hdl := entity.GetHdl()
	if _, exists := m.entities[hdl]; exists {
		return fmt.Errorf("entity already registered: hdl=%d", hdl)
	}

	m.entities[hdl] = entity
	return nil
}

// Unregister 注销实体
func (m *EntityMgr) Unregister(hdl uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entities, hdl)
}

// GetByHdl 通过hdl获取实体
func (m *EntityMgr) GetByHdl(hdl uint64) (iface.IEntity, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entity, ok := m.entities[hdl]
	return entity, ok
}

// GetById 通过Id获取实体列表（可能有多个相同Id的实体）
func (m *EntityMgr) GetById(Id uint64) []iface.IEntity {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entities := make([]iface.IEntity, 0)
	for _, entity := range m.entities {
		if entity.GetId() == Id {
			entities = append(entities, entity)
		}
	}
	return entities
}

// GetAll 获取所有实体
func (m *EntityMgr) GetAll() []iface.IEntity {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entities := make([]iface.IEntity, 0, len(m.entities))
	for _, entity := range m.entities {
		entities = append(entities, entity)
	}
	return entities
}

// GetCount 获取实体数量
func (m *EntityMgr) GetCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entities)
}

// Clear 清空所有实体
func (m *EntityMgr) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entities = make(map[uint64]iface.IEntity)
}
