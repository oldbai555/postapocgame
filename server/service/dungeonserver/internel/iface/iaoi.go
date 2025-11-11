/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

type IAOISys interface {
	GetVisibleEntities() []IEntity
	AddVisibleEntity(entity IEntity)
	RemoveVisibleEntity(entityId uint64)
	IsVisible(entityId uint64) bool
	ClearVisibleEntities()
}
