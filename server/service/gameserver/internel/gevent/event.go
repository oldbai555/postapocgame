package gevent

import (
	"context"
	"postapocgame/server/internal/event"
	"sync"
)

// 服务器级别的事件总线（用于服务器启动等全局事件）
var eventBus = event.NewEventBus()

type playerEventRegistration struct {
	eventType event.Type
	priority  int
	handler   event.Handler
}

var (
	playerEventRegistrations []playerEventRegistration
	playerRegMu              sync.RWMutex
)

// Subscribe 订阅服务器级别事件
func Subscribe(eventType event.Type, handler event.Handler) {
	eventBus.Subscribe(eventType, 0, handler)
}

// Publish 发布服务器级别事件
func Publish(ctx context.Context, event *event.Event) {
	eventBus.Publish(ctx, event)
}

// registerPlayerEvent 记录玩家事件的订阅器，稍后在每个玩家的事件总线上复用
func registerPlayerEvent(eventType event.Type, priority int, handler event.Handler) {
	if handler == nil {
		return
	}
	playerRegMu.Lock()
	defer playerRegMu.Unlock()
	playerEventRegistrations = append(playerEventRegistrations, playerEventRegistration{
		eventType: eventType,
		priority:  priority,
		handler:   handler,
	})
}

// SubscribePlayerEventH 订阅玩家事件（高优先级）
func SubscribePlayerEventH(eventType event.Type, handler event.Handler) {
	registerPlayerEvent(eventType, 3, handler)
}

// SubscribePlayerEvent 订阅玩家事件（默认优先级）
func SubscribePlayerEvent(eventType event.Type, handler event.Handler) {
	registerPlayerEvent(eventType, 2, handler)
}

// SubscribePlayerEventL 订阅玩家事件（低优先级）
func SubscribePlayerEventL(eventType event.Type, handler event.Handler) {
	registerPlayerEvent(eventType, 1, handler)
}

// NewPlayerEventBus 为单个玩家创建事件总线，并应用预注册的订阅器
func NewPlayerEventBus() *event.Bus {
	bus := event.NewEventBus()
	playerRegMu.RLock()
	defer playerRegMu.RUnlock()
	for _, reg := range playerEventRegistrations {
		bus.Subscribe(reg.eventType, reg.priority, reg.handler)
	}
	return bus
}
