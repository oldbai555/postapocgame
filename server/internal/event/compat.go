package event

import (
	"context"
	"sync"
)

// 兼容旧接口的包装层

var (
	compatOnce     sync.Once
	compatEventBus *Bus
)

// getCompatBus 获取兼容模式的事件总线
func getCompatBus() *Bus {
	compatOnce.Do(func() {
		compatEventBus = NewEventBus()
	})
	return compatEventBus
}

// Subscribe 订阅事件
func Subscribe(eventType Type, handler func(event *Event) error) {
	bus := getCompatBus()

	// 包装旧的 handler 为新接口
	newHandler := func(ctx context.Context, event *Event) error {
		return handler(event)
	}

	bus.Subscribe(eventType, 0, newHandler)
}

// Publish 发布事件
func Publish(typ Type, args ...interface{}) error {
	bus := getCompatBus()

	event := &Event{
		Type: typ,
		Data: args,
	}

	return bus.Publish(context.Background(), event)
}

// GetEventBus 获取全局事件总线（兼容旧接口）
// 注意：这个返回的是兼容模式的 Bus，不是新的 actor-based 系统
func GetEventBus() *Bus {
	return getCompatBus()
}

// SubscribeWithPriority 带优先级订阅（兼容旧接口，扩展功能）
func SubscribeWithPriority(eventType Type, priority int, handler func(event *Event) error) {
	bus := getCompatBus()

	newHandler := func(ctx context.Context, event *Event) error {
		return handler(event)
	}

	bus.Subscribe(eventType, priority, newHandler)
}
