package iface

import (
	"context"
	"postapocgame/server/internal/actor"
)

// PublicActorGateway PublicActor 交互接口（Use Case 层定义）
type PublicActorGateway interface {
	// SendMessageAsync 发送异步消息到 PublicActor
	SendMessageAsync(ctx context.Context, key string, message actor.IActorMessage) error

	// RegisterHandler 注册消息处理器（可选，用于反向调用）
	RegisterHandler(msgId uint16, handler actor.HandlerMessageFunc)
}
