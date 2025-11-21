/**
 * @Author: zjj
 * @Date: 2025/11/13
 * @Desc:
**/

package gshare

import "postapocgame/server/internal/servertime"

var (
	platformId     uint32
	srvId          uint32
	openSrvTimeSec int64
)

func SetPlatformId(id uint32) {
	platformId = id
}
func SetSrvId(id uint32) {
	srvId = id
}
func GetPlatformId() uint32 {
	return platformId
}
func GetSrvId() uint32 {
	return srvId
}

func SetOpenSrvTime(openTimeSec int64) {
	openSrvTimeSec = openTimeSec
}

func GetOpenSrvUnix() int64 {
	return openSrvTimeSec
}

// GetOpenSrvDay 返回开服第几天（从1开始），若未设置开服时间返回0
func GetOpenSrvDay() int64 {
	if openSrvTimeSec <= 0 {
		return 0
	}
	return servertime.DaysSince(openSrvTimeSec)
}
