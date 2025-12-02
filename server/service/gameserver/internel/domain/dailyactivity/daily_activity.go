package dailyactivity

import (
	"time"

	"postapocgame/server/internal/protocol"
)

// EnsureData 确保 DailyActivityData 初始化
func EnsureData(binaryData *protocol.PlayerRoleBinaryData) *protocol.SiDailyActivityData {
	if binaryData == nil {
		return nil
	}
	if binaryData.DailyActivityData == nil {
		binaryData.DailyActivityData = &protocol.SiDailyActivityData{
			RewardStates: make([]*protocol.DailyActivityRewardState, 0),
		}
	}
	if binaryData.DailyActivityData.RewardStates == nil {
		binaryData.DailyActivityData.RewardStates = make([]*protocol.DailyActivityRewardState, 0)
	}
	return binaryData.DailyActivityData
}

// NeedReset 判断是否需要按天重置
func NeedReset(data *protocol.SiDailyActivityData, now time.Time) bool {
	if data == nil {
		return false
	}
	if data.LastResetTime == 0 {
		return true
	}
	last := time.Unix(data.LastResetTime, 0).In(time.Local)
	now = now.In(time.Local)
	return last.Year() != now.Year() || last.YearDay() != now.YearDay()
}

// ResetForNewDay 执行每日重置
func ResetForNewDay(data *protocol.SiDailyActivityData, now time.Time) {
	if data == nil {
		return
	}
	data.TodayPoints = 0
	data.LastResetTime = now.Unix()
	for _, st := range data.RewardStates {
		if st != nil {
			st.Claimed = false
		}
	}
}
