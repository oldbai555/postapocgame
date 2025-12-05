package interfaces

import (
	"context"
	"postapocgame/server/internal/event"
)

// EventPublisher 事件发布器接口（Use Case 层定义）
type EventPublisher interface {
	// PublishPlayerEvent 发布玩家事件
	PublishPlayerEvent(ctx context.Context, eventType event.Type, args ...interface{})

	// SubscribePlayerEvent 订阅玩家事件
	SubscribePlayerEvent(eventType event.Type, handler func(ctx context.Context, event interface{}))
}
