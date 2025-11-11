/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

type IBuffSys interface {
	AddBuff(entityId uint64, buffId uint32, casterId uint64) error
	RemoveBuff(entityId uint64, buffId uint32) error
	HasBuff(entityId uint64, buffId uint32) bool
}
