package actor

import (
	"context"
)

type ActorMode int

const (
	// ModeSingle 单例模式 - 所有消息在同一个Actor中处理
	ModeSingle ActorMode = iota

	// ModePerKey 每Key模式 - 每个Key独立的Actor
	ModePerKey
)

type IActorContext interface {
	// GetID 获取Actor的唯一标识
	GetID() string

	// ExecuteAsync 在Actor上下文中异步执行
	ExecuteAsync(message IActorMessage)

	// GetData 获取绑定的数据
	GetData(key string) interface{}

	// SetData 设置绑定的数据
	SetData(key string, data interface{})

	// IsRunning 是否正在运行
	IsRunning() bool
}

type IActorManager interface {
	GetOrCreateActor(key string) (IActorContext, error)
	RemoveActor(key string) error

	SendMessageAsync(key string, message IActorMessage) error
	BroadcastAsync(message IActorMessage)

	Init() error

	// Start 启动管理器
	Start(ctx context.Context) error

	// Stop 停止管理器
	Stop(ctx context.Context) error

	// GetMode 获取运行模式
	GetMode() ActorMode
}

type IActorHandler interface {
	RegisterMessageHandler(msgId uint16, f HandlerMessageFunc)
	HandleMessage(msg IActorMessage)

	Loop()
	OnInit()
	OnStart()
	OnStop()

	SetActorContext(IActorContext)
}

type IActorMessage interface {
	GetMsgId() uint16
	GetData() []byte
	GetContext() context.Context
}
