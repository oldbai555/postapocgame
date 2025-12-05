package interfaces

import "time"

// ChatRateLimiter 聊天冷却接口
type ChatRateLimiter interface {
	CanSend(now time.Time, cooldown time.Duration) bool
	MarkSent(now time.Time)
}
