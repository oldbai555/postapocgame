/**
 * @Author: zjj
 * @Date: 2025/11/25
 * @Desc:
**/

package iface

type IDrop interface {
	IEntity
	IsOwner(entity IEntity) bool
	GetItemId() uint32
	GetCount() uint32
	GetOwnerHdl() uint64
	GetOwnerRoleId() uint64
}
