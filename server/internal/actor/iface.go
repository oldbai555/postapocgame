package actor

import (
	"context"
)

// IActor Actor接口
type IActor interface {
	// Start 启动Actor
	Start(ctx context.Context) error
	// Stop 停止Actor
	Stop(ctx context.Context) error
	// Send 发送消息到Actor(异步)
	Send(msg *Message) error
	// Call 调用Actor(同步,等待响应)
	Call(ctx context.Context, msg *Message) (*Message, error)
	GetMode() Mode
}

// IActorSystem Actor系统接口
type IActorSystem interface {
	Init()
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	Send(sessionId string, msgId uint16, data []byte) error
}

type IActorMsgHandler interface {
	HandleActorMessage(msg *Message) error
	Loop()
	OnInit()
}
