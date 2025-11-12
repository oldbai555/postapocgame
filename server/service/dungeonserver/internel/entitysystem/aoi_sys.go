package entitysystem

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
)

// AOI 观察者系统 (Area Of Interest)
// 使用九宫格算法
type AOI struct {
	entity          iface.IEntity
	visibleEntities map[uint64]iface.IEntity // 可见的实体列表
	mu              sync.RWMutex
}

// NewAOI 创建AOI
func NewAOI(entity iface.IEntity) *AOI {
	return &AOI{
		entity:          entity,
		visibleEntities: make(map[uint64]iface.IEntity),
	}
}

// GetVisibleEntities 获取可见实体列表
func (aoi *AOI) GetVisibleEntities() []iface.IEntity {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	entities := make([]iface.IEntity, 0, len(aoi.visibleEntities))
	for _, entity := range aoi.visibleEntities {
		entities = append(entities, entity)
	}
	return entities
}

// AddVisibleEntity 添加可见实体
func (aoi *AOI) AddVisibleEntity(entity iface.IEntity) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	aoi.visibleEntities[entity.GetId()] = entity

	// TODO: 通知客户端有新实体进入视野
}

// RemoveVisibleEntity 移除可见实体
func (aoi *AOI) RemoveVisibleEntity(entityId uint64) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	delete(aoi.visibleEntities, entityId)

	// TODO: 通知客户端有实体离开视野
}

// IsVisible 检查实体是否在视野内
func (aoi *AOI) IsVisible(entityId uint64) bool {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	_, ok := aoi.visibleEntities[entityId]
	return ok
}

// ClearVisibleEntities 清空可见实体
func (aoi *AOI) ClearVisibleEntities() {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	aoi.visibleEntities = make(map[uint64]iface.IEntity)
}

// OnMove 实体移动时更新AOI
func (aoi *AOI) OnMove(oldPos, newPos *argsdef.Position) {
	// 获取旧的九宫格
	oldGrIds := argsdef.GetNineGrIds(oldPos)
	// 获取新的九宫格
	newGrIds := argsdef.GetNineGrIds(newPos)

	// 找出离开的格子
	for _, oldGrId := range oldGrIds {
		found := false
		for _, newGrId := range newGrIds {
			if oldGrId == newGrId {
				found = true
				break
			}
		}
		if !found {
			// 格子离开，需要移除该格子内的实体
			// TODO: 实现
		}
	}

	// 找出进入的格子
	for _, newGrId := range newGrIds {
		found := false
		for _, oldGrId := range oldGrIds {
			if newGrId == oldGrId {
				found = true
				break
			}
		}
		if !found {
			// 格子进入，需要添加该格子内的实体
			// TODO: 实现
		}
	}
}
