package servertime

import (
	"sync/atomic"
	"time"
)

type timeFunc func() time.Time

var (
	currentProvider atomic.Value
	offsetMillis    atomic.Int64
)

func init() {
	currentProvider.Store(defaultTimeProvider)
}

func defaultTimeProvider() time.Time {
	return time.Now().UTC()
}

// Now 返回统一的服务器时间（UTC）
func Now() time.Time {
	now := currentProvider.Load().(timeFunc)()
	ms := offsetMillis.Load()
	if ms == 0 {
		return now
	}
	return now.Add(time.Duration(ms) * time.Millisecond)
}

// UnixMilli 返回当前服务器时间的 Unix 毫秒
func UnixMilli() int64 {
	return Now().UnixMilli()
}

// Since 返回给定时间点到当前服务器时间的间隔
func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}

// SetTimeProvider 替换时间源（测试用途）
func SetTimeProvider(fn func() time.Time) {
	if fn == nil {
		currentProvider.Store(defaultTimeProvider)
		return
	}
	currentProvider.Store(timeFunc(func() time.Time {
		return fn().UTC()
	}))
}

// AddOffset 为服务器时间增加偏移（可配置NTP校准等场景），单位：毫秒
func AddOffset(delta time.Duration) {
	offsetMillis.Add(delta.Milliseconds())
}

// ResetOffset 清空额外时间偏移
func ResetOffset() {
	offsetMillis.Store(0)
}
