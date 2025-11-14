package entitysystem

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
)

type AOISys struct {
	entity          iface.IEntity
	visibleEntities map[uint64]iface.IEntity // 可见的实体列表
	mu              sync.RWMutex
	pendingEnter    []iface.IEntity
	pendingLeave    []uint64
}

func NewAOISys(entity iface.IEntity) *AOISys {
	return &AOISys{
		entity:          entity,
		visibleEntities: make(map[uint64]iface.IEntity),
	}
}

// GetVisibleEntities 获取可见实体列表
func (aoi *AOISys) GetVisibleEntities() []iface.IEntity {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	entities := make([]iface.IEntity, 0, len(aoi.visibleEntities))
	for _, entity := range aoi.visibleEntities {
		entities = append(entities, entity)
	}
	return entities
}

// AddVisibleEntity 添加可见实体
func (aoi *AOISys) AddVisibleEntity(entity iface.IEntity) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	aoi.visibleEntities[entity.GetId()] = entity
	aoi.pendingEnter = append(aoi.pendingEnter, entity)
}

// RemoveVisibleEntity 移除可见实体
func (aoi *AOISys) RemoveVisibleEntity(entityId uint64) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	delete(aoi.visibleEntities, entityId)
	aoi.pendingLeave = append(aoi.pendingLeave, entityId)
}

// IsVisible 检查实体是否在视野内
func (aoi *AOISys) IsVisible(entityId uint64) bool {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	_, ok := aoi.visibleEntities[entityId]
	return ok
}

// ClearVisibleEntities 清空可见实体
func (aoi *AOISys) ClearVisibleEntities() {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	aoi.visibleEntities = make(map[uint64]iface.IEntity)
	aoi.pendingEnter = nil
	aoi.pendingLeave = nil
}

// OnMove 实体移动时更新AOI
func (aoi *AOISys) OnMove(oldPos, newPos *argsdef.Position) {
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

// ConsumeVisibilityChanges 获取一次性视野变化
func (aoi *AOISys) ConsumeVisibilityChanges() (enter []iface.IEntity, leave []uint64) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	if len(aoi.pendingEnter) > 0 {
		enter = append(enter, aoi.pendingEnter...)
		aoi.pendingEnter = nil
	}
	if len(aoi.pendingLeave) > 0 {
		leave = append(leave, aoi.pendingLeave...)
		aoi.pendingLeave = nil
	}
	return
}
