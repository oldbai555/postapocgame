package entitymgr

import (
	"fmt"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"time"
)

// EntityMgr 全局实体管理器
type EntityMgr struct {
	entities     map[uint64]iface.IEntity // hdl -> entity
	sessions     map[string]uint64        // sessionId -> entity hdl
	entityScenes map[uint64]iface.IScene
}

var (
	globalEntityMgr *EntityMgr
)

// GetEntityMgr 获取全局实体管理器
func GetEntityMgr() *EntityMgr {
	if globalEntityMgr == nil {
		globalEntityMgr = &EntityMgr{
			entities:     make(map[uint64]iface.IEntity),
			sessions:     make(map[string]uint64),
			entityScenes: make(map[uint64]iface.IScene),
		}
	}
	return globalEntityMgr
}

// Register 注册实体
func (m *EntityMgr) Register(entity iface.IEntity) error {
	hdl := entity.GetHdl()
	if _, exists := m.entities[hdl]; exists {
		return fmt.Errorf("entity already registered: hdl=%d", hdl)
	}

	m.entities[hdl] = entity
	return nil
}

// Unregister 注销实体
func (m *EntityMgr) Unregister(hdl uint64) {
	delete(m.entities, hdl)
}

// BindSession 绑定sessionId与实体
func (m *EntityMgr) BindSession(sessionId string, hdl uint64) {
	if sessionId == "" {
		return
	}
	m.sessions[sessionId] = hdl
}

// UnbindSession 解除session绑定
func (m *EntityMgr) UnbindSession(sessionId string) {
	if sessionId == "" {
		return
	}
	delete(m.sessions, sessionId)
}

// GetBySession 根据sessionId获取实体
func (m *EntityMgr) GetBySession(sessionId string) (iface.IEntity, bool) {
	if sessionId == "" {
		return nil, false
	}
	hdl, ok := m.sessions[sessionId]
	if !ok {
		return nil, false
	}
	return m.GetByHdl(hdl)
}

// BindScene 绑定实体所在场景
func (m *EntityMgr) BindScene(hdl uint64, scene iface.IScene) {
	m.entityScenes[hdl] = scene
}

// UnbindScene 解除实体与场景绑定
func (m *EntityMgr) UnbindScene(hdl uint64) {
	delete(m.entityScenes, hdl)
}

// GetSceneByHandle 获取实体所在场景
func (m *EntityMgr) GetSceneByHandle(hdl uint64) (iface.IScene, bool) {
	scene, ok := m.entityScenes[hdl]
	return scene, ok
}

// GetByHdl 通过hdl获取实体
func (m *EntityMgr) GetByHdl(hdl uint64) (iface.IEntity, bool) {
	entity, ok := m.entities[hdl]
	return entity, ok
}

// GetById 通过Id获取实体列表（可能有多个相同Id的实体）
func (m *EntityMgr) GetById(Id uint64) []iface.IEntity {
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
	entities := make([]iface.IEntity, 0, len(m.entities))
	for _, entity := range m.entities {
		entities = append(entities, entity)
	}
	return entities
}

// RunOne 执行所有实体的单帧逻辑
func (m *EntityMgr) RunOne(now time.Time) {
	entities := m.GetAll()
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		entity.RunOne(now)
	}
}

// RunOne 遍历所有实体并执行单帧逻辑（全局入口）
func RunOne(now time.Time) {
	GetEntityMgr().RunOne(now)
}

// GetCount 获取实体数量
func (m *EntityMgr) GetCount() int {
	return len(m.entities)
}

// Clear 清空所有实体
func (m *EntityMgr) Clear() {
	m.entities = make(map[uint64]iface.IEntity)
	m.entityScenes = make(map[uint64]iface.IScene)
}
