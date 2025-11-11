package gevent

import (
	"context"
	"postapocgame/server/internal/event"
)

// 服务器级别的事件总线（用于服务器启动等全局事件）
var eventBus = event.NewEventBus()

// 玩家级别的事件总线模板（用于克隆）
var playerEventBusTemplate = event.NewEventBus()

// Subscribe 订阅服务器级别事件
func Subscribe(eventType event.Type, handler event.Handler) {
	eventBus.Subscribe(eventType, 0, handler)
}

// Publish 发布服务器级别事件
func Publish(ctx context.Context, event *event.Event) {
	eventBus.Publish(ctx, event)
}

func SubscribePlayerEventH(eventType event.Type, handler event.Handler) {
	playerEventBusTemplate.Subscribe(eventType, 3, handler)
}

func SubscribePlayerEvent(eventType event.Type, handler event.Handler) {
	playerEventBusTemplate.Subscribe(eventType, 2, handler)
}

func SubscribePlayerEventL(eventType event.Type, handler event.Handler) {
	playerEventBusTemplate.Subscribe(eventType, 1, handler)
}

// ClonePlayerEventBus 克隆玩家事件总线（为新玩家创建独立的事件总线）
func ClonePlayerEventBus() *event.Bus {
	return playerEventBusTemplate.CloneByReplay()
}
