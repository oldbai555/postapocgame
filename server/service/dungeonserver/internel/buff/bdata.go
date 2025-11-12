/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package buff

import (
	"postapocgame/server/internal/jsonconf"
	"time"
)

type BData struct {
	BuffId     uint32
	BuffName   string
	BuffType   uint32
	StackCount uint32
	MaxStack   uint32
	Duration   time.Duration
	StartTime  time.Time
	EndTime    time.Time
	CasterId   uint64 // 施法者Id

	// Buff效果
	Effects []jsonconf.BuffEffect
}
