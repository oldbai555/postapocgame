/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

type IScene interface {
	AddEntity(IEntity) error
	RemoveEntity(hdl uint64) error
	GetEntity(hdl uint64) (IEntity, bool)
	GetAllEntities() []IEntity
	EntityMove(hdl uint64, newX, newY uint32) error
	IsWalkable(x, y int) bool
	GetRandomWalkablePos() (uint32, uint32)

	GetSceneId() uint32
	GetFuBenId() uint32
	GetFuBen() IFuBen
}
