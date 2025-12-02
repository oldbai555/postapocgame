package interfaces

// EventSubscriber 事件订阅接口（Use Case 层，可选）
type EventSubscriber interface {
	// SubscribeEvents 订阅事件
	SubscribeEvents(publisher EventPublisher)
}
