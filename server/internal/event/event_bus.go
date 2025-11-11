package event

import (
	"context"
	"sort"
	"sync"
)

// Handler 事件处理器（支持 context）
type Handler func(ctx context.Context, event *Event) error

// handlerEntry 带优先级的处理器条目
type handlerEntry struct {
	priority int
	handler  Handler
}

// Bus 事件总线（支持优先级和 registry 机制）
type Bus struct {
	mu          sync.RWMutex
	subscribers map[Type][]handlerEntry

	// registry 存储"重放"函数，用于克隆/重建
	registry []func(b *Bus)
}

// NewEventBus 创建事件总线
func NewEventBus() *Bus {
	return &Bus{
		subscribers: make(map[Type][]handlerEntry),
		registry:    make([]func(b *Bus), 0, 32),
	}
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

	// 保存到 registry 用于克隆
	eb.registry = append(eb.registry, func(b *Bus) {
		b.Subscribe(eventType, priority, handler)
	})
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
func (eb *Bus) Publish(ctx context.Context, event *Event) error {
	eb.mu.RLock()
	handlers := eb.subscribers[event.Type]
	eb.mu.RUnlock()

	// 按已排序的优先级顺序执行
	for _, he := range handlers {
		if err := he.handler(ctx, event); err != nil {
			// 可以选择：
			// 1. 继续执行其他 handlers
			// 2. 中断并返回错误
			// 这里选择继续执行，但记录第一个错误
			return err
		}
	}

	return nil
}

// PublishAsync 异步发布事件（每个 handler 在独立 goroutine 中执行）
func (eb *Bus) PublishAsync(ctx context.Context, event *Event) {
	eb.mu.RLock()
	handlers := append([]handlerEntry(nil), eb.subscribers[event.Type]...)
	eb.mu.RUnlock()

	var wg sync.WaitGroup
	for _, he := range handlers {
		wg.Add(1)
		go func(h Handler) {
			defer wg.Done()
			_ = h(ctx, event)
		}(he.handler)
	}
	wg.Wait()
}

// CloneByReplay 通过 registry 重建一个新的 EventBus
// 用于为每个 actor 创建独立的 localBus
func (eb *Bus) CloneByReplay() *Bus {
	newBus := NewEventBus()

	eb.mu.RLock()
	regs := append([]func(b *Bus){}, eb.registry...)
	eb.mu.RUnlock()

	for _, reg := range regs {
		reg(newBus)
	}

	return newBus
}

// Clear 清空所有订阅（用于测试或重置）
func (eb *Bus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers = make(map[Type][]handlerEntry)
	eb.registry = make([]func(b *Bus), 0)
}

// GetSubscriberCount 获取某个事件的订阅者数量
func (eb *Bus) GetSubscriberCount(eventType Type) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return len(eb.subscribers[eventType])
}
