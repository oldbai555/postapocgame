package event

import (
	"context"
	"postapocgame/server/pkg/routine"
	"sort"
	"sync"
)

func NewEventBus() *Bus {
	return &Bus{
		subscribers: make(map[Type][]handlerEntry),
	}
}

type Bus struct {
	mu          sync.RWMutex
	subscribers map[Type][]handlerEntry
}

type handlerEntry struct {
	priority int
	handler  Handler
}

// Subscribe 订阅事件（支持优先级）
// priority: 数值越大优先级越高，先执行
func (eb *Bus) Subscribe(eventType Type, priority int, handler Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	entry := handlerEntry{
		priority: priority,
		handler:  handler,
	}

	list := eb.subscribers[eventType]
	list = append(list, entry)

	// 按优先级排序（高优先级在前）
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].priority > list[j].priority
	})

	eb.subscribers[eventType] = list
}

// SubscribeWithDefaultPriority 使用默认优先级订阅
func (eb *Bus) SubscribeWithDefaultPriority(eventType Type, handler Handler) {
	eb.Subscribe(eventType, 0, handler)
}

// Unsubscribe 取消订阅
func (eb *Bus) Unsubscribe(eventType Type) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	delete(eb.subscribers, eventType)
}

// Publish 发布事件（同步执行所有 handlers）
// 按照优先级顺序执行，保证确定性
func (eb *Bus) Publish(ctx context.Context, event *Event) {
	eb.mu.RLock()
	handlers := eb.subscribers[event.Type]
	eb.mu.RUnlock()

	// 按已排序的优先级顺序执行
	for _, he := range handlers {
		routine.Run(func() {
			he.handler(ctx, event)
		})
	}

	return
}

// CloneByReplay 通过 registry 重建一个新的 EventBus
// 用于为每个 actor 创建独立的 localBus
func (eb *Bus) CloneByReplay() *Bus {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	newBus := NewEventBus()

	newBus.subscribers = make(map[Type][]handlerEntry, len(eb.subscribers))
	for typ, val := range eb.subscribers {
		var list = val
		var cpList = make([]handlerEntry, len(list))
		copy(list, cpList)
		newBus.subscribers[typ] = cpList
	}
	return newBus
}
