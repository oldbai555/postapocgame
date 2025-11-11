package gevent

import (
	"context"
	"postapocgame/server/internal/event"
)

var eventBus = event.NewEventBus()

func Subscribe(eventType event.Type, handler event.Handler) {
	eventBus.Subscribe(eventType, 0, handler)
}
func Publish(ctx context.Context, event *event.Event) {
	eventBus.Publish(ctx, event)
}
