package event

import (
	"context"
	"postapocgame/server/internal/event"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// EventAdapter 事件适配器（实现 Use Case 层的 EventPublisher 接口）
type EventAdapter struct{}

// NewEventAdapter 创建事件适配器
func NewEventAdapter() interfaces.EventPublisher {
	return &EventAdapter{}
}

// PublishPlayerEvent 发布玩家事件
func (a *EventAdapter) PublishPlayerEvent(ctx context.Context, eventType event.Type, args ...interface{}) {
	playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return
	}
	// 使用玩家自己的事件总线发布事件
	playerRole.Publish(eventType, args...)
}

// SubscribePlayerEvent 订阅玩家事件
// 注意：此方法在系统初始化时调用，订阅到全局模板，新玩家会自动继承
func (a *EventAdapter) SubscribePlayerEvent(eventType event.Type, handler func(ctx context.Context, event interface{})) {
	// 订阅到全局模板，新玩家会自动继承
	gevent.SubscribePlayerEvent(eventType, func(evCtx context.Context, ev *event.Event) {
		handler(evCtx, ev)
	})
}
