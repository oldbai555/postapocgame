package scene

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
)

// AOIManager AOI管理器（场景级别）
type AOIManager struct {
	mu    sync.RWMutex
	grIds map[argsdef.GrIdSt]map[uint64]iface.IEntity // grIdId -> entityId -> entity
}

// NewAOIManager 创建AOI管理器
func NewAOIManager() *AOIManager {
	return &AOIManager{
		grIds: make(map[argsdef.GrIdSt]map[uint64]iface.IEntity),
	}
}

// AddEntity 添加实体到AOI管理器
func (am *AOIManager) AddEntity(entity iface.IEntity) {
	am.mu.Lock()
	defer am.mu.Unlock()

	pos := entity.GetPosition()
	grIds := argsdef.GetNineGrIds(pos)

	for _, grId := range grIds {
		if _, ok := am.grIds[grId]; !ok {
			am.grIds[grId] = make(map[uint64]iface.IEntity)
		}
		am.grIds[grId][entity.GetId()] = entity
	}

	// 更新该实体的可见列表
	am.updateEntityVisibility(entity)
}

// RemoveEntity 从AOI管理器移除实体
func (am *AOIManager) RemoveEntity(entity iface.IEntity) {
	am.mu.Lock()
	defer am.mu.Unlock()

	pos := entity.GetPosition()
	grIds := argsdef.GetNineGrIds(pos)

	for _, grId := range grIds {
		if entities, ok := am.grIds[grId]; ok {
			delete(entities, entity.GetId())
			if len(entities) == 0 {
				delete(am.grIds, grId)
			}
		}
	}

	// 清空可见列表
	entity.GetAOISys().ClearVisibleEntities()
}

// UpdateEntity 更新实体位置
func (am *AOIManager) UpdateEntity(entity iface.IEntity, oldPos, newPos *argsdef.Position) {
	oldGrIds := argsdef.GetNineGrIds(oldPos)
	newGrIds := argsdef.GetNineGrIds(newPos)

	am.mu.Lock()

	// 找出需要离开的格子
	for _, oldGrId := range oldGrIds {
		found := false
		for _, newGrId := range newGrIds {
			if oldGrId == newGrId {
				found = true
				break
			}
		}
		if !found {
			if entities, ok := am.grIds[oldGrId]; ok {
				delete(entities, entity.GetId())
				if len(entities) == 0 {
					delete(am.grIds, oldGrId)
				}
			}
		}
	}

	// 找出需要进入的格子
	for _, newGrId := range newGrIds {
		found := false
		for _, oldGrId := range oldGrIds {
			if newGrId == oldGrId {
				found = true
				break
			}
		}
		if !found {
			if _, ok := am.grIds[newGrId]; !ok {
				am.grIds[newGrId] = make(map[uint64]iface.IEntity)
			}
			am.grIds[newGrId][entity.GetId()] = entity
		}
	}

	am.mu.Unlock() // 🔧 提前释放锁

	// 🔧 在锁外更新可见列表（避免持有锁时间过长）
	am.updateEntityVisibility(entity)
}

// updateEntityVisibility 更新实体的可见列表
func (am *AOIManager) updateEntityVisibility(entity iface.IEntity) {
	pos := entity.GetPosition()
	grIds := argsdef.GetNineGrIds(pos)

	// 收集九宫格内的所有实体
	visibleEntities := make(map[uint64]iface.IEntity)
	for _, grId := range grIds {
		if entities, ok := am.grIds[grId]; ok {
			for Id, e := range entities {
				// 不包括自己
				if Id != entity.GetId() {
					visibleEntities[Id] = e
				}
			}
		}
	}

	aoi := entity.GetAOISys()

	// 找出新进入视野的实体
	for Id, e := range visibleEntities {
		if !aoi.IsVisible(Id) {
			aoi.AddVisibleEntity(e)
		}
	}

	// 找出离开视野的实体
	currentVisible := aoi.GetVisibleEntities()
	for _, e := range currentVisible {
		if _, ok := visibleEntities[e.GetId()]; !ok {
			aoi.RemoveVisibleEntity(e.GetId())
		}
	}
}

// GetEntitiesInRange 获取范围内的所有实体
func (am *AOIManager) GetEntitiesInRange(pos *argsdef.Position) []iface.IEntity {
	am.mu.RLock()
	defer am.mu.RUnlock()

	grIds := argsdef.GetNineGrIds(pos)

	entitiesMap := make(map[uint64]iface.IEntity)
	for _, grId := range grIds {
		if entities, ok := am.grIds[grId]; ok {
			for Id, entity := range entities {
				entitiesMap[Id] = entity
			}
		}
	}

	result := make([]iface.IEntity, 0, len(entitiesMap))
	for _, entity := range entitiesMap {
		result = append(result, entity)
	}

	return result
}
