/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

import "time"

type IBuffSys interface {
	AddBuff(buffId uint32, caster IEntity) error
	RemoveBuff(buffId uint32) error
	HasBuff(buffId uint32) bool
	RunOne(now time.Time)
}
