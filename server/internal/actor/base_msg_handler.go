package actor

import (
	"fmt"
	"sync"
)

type MsgHandlerFunc func(msg *Message) error

var _ IActorMsgHandler = (*BaseActorMsgHandler)(nil)

type BaseActorMsgHandler struct {
	handlers map[uint16]MsgHandlerFunc // msgId -> 处理函数
	mu       sync.RWMutex
}

// Register 注册消息处理函数
func (b *BaseActorMsgHandler) Register(msgId uint16, handler MsgHandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[msgId] = handler
}

// Unregister 取消注册消息处理函数
func (b *BaseActorMsgHandler) Unregister(msgId uint16) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.handlers, msgId)
}

func (b *BaseActorMsgHandler) HandleActorMessage(msg *Message) error {
	b.mu.RLock()
	handler, ok := b.handlers[msg.MsgId]
	b.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no handler registered for msgId: %d", msg.MsgId)
	}

	return handler(msg)
}

// Loop Actor循环处理(每次处理消息前调用)
func (b *BaseActorMsgHandler) Loop() {
	// 子类可以重写此方法,实现自定义的循环逻辑
}

// OnInit 初始化(子类重写此方法注册消息处理函数)
func (b *BaseActorMsgHandler) OnInit() {
	// 子类重写此方法,注册消息处理函数
	if b.handlers == nil {
		b.handlers = make(map[uint16]MsgHandlerFunc)
	}
}
