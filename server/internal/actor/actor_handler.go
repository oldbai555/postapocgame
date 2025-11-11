/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package actor

import (
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"sync"
)

var _ IActorHandler = (*BaseActorHandler)(nil)

type HandlerMessageFunc func(msg IActorMessage)

type BaseActorHandler struct {
	rw           sync.RWMutex
	skipRegistry bool
	handlerMap   map[uint16]HandlerMessageFunc
	registry     []func(nb *BaseActorHandler)
}

func NewBaseActorHandler() *BaseActorHandler {
	return &BaseActorHandler{
		handlerMap: make(map[uint16]HandlerMessageFunc),
	}
}

func (b *BaseActorHandler) RegisterMessageHandler(msgId uint16, f HandlerMessageFunc) {
	b.rw.Lock()
	defer b.rw.Unlock()

	if _, ok := b.handlerMap[msgId]; ok {
		log.Fatalf("msgId %d already register", msgId)
	}
	b.handlerMap[msgId] = f

	if b.skipRegistry {
		return
	}

	// ✅ 修复闭包捕获问题
	mid := msgId
	fn := f
	b.registry = append(b.registry, func(nb *BaseActorHandler) {
		nb.RegisterMessageHandler(mid, fn)
	})
}

func (b *BaseActorHandler) HandleMessage(msg IActorMessage) {
	b.rw.RLock()
	defer b.rw.RLock()
	f := b.handlerMap[msg.GetMsgId()]
	if f == nil {
		log.Errorf("msgId %d not found handler", msg.GetMsgId())
		return
	}
	routine.Run(func() {
		f(msg)
	})
	return
}

func (b *BaseActorHandler) Loop() {}

func (b *BaseActorHandler) OnInit() {
	if b.handlerMap == nil {
		b.handlerMap = make(map[uint16]HandlerMessageFunc)
	}
}

func (b *BaseActorHandler) OnStart() {}

func (b *BaseActorHandler) OnStop() {}

func (b *BaseActorHandler) Clone() *BaseActorHandler {
	b.rw.Lock()
	defer b.rw.Unlock()
	handler := NewBaseActorHandler()
	for _, f := range b.registry {
		f(handler)
	}
	handler.skipRegistry = true
	return handler
}
