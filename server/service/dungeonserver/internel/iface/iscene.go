/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

type IScene interface {
	GetEntity(hdl uint64) (IEntity, bool)
	GetAllEntities() []IEntity

	GetSceneId() uint32
	GetFuBenId() uint32
}
