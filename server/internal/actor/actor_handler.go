/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package actor

import (
	"postapocgame/server/pkg/log"
	"sync"
)

var _ IActorHandler = (*BaseActorHandler)(nil)

type HandlerMessageFunc func(msg IActorMessage)

type BaseActorHandler struct {
	rw         sync.RWMutex
	handlerMap map[uint16]HandlerMessageFunc
	name       string
}

func NewBaseActorHandler(name string) *BaseActorHandler {
	return &BaseActorHandler{
		handlerMap: make(map[uint16]HandlerMessageFunc),
		name:       name,
	}
}

func (b *BaseActorHandler) RegisterMessageHandler(msgId uint16, f HandlerMessageFunc) {
	b.rw.Lock()
	defer b.rw.Unlock()

	if _, ok := b.handlerMap[msgId]; ok {
		log.Fatalf("[%s] msgId %d already register", b.name, msgId)
	}
	b.handlerMap[msgId] = f
}

func (b *BaseActorHandler) HandleMessage(msg IActorMessage) {
	b.rw.RLock()
	f := b.handlerMap[msg.GetMsgId()]
	b.rw.RUnlock() // ✅ 修复：RLock 后应该 RUnlock

	if f == nil {
		log.Errorf("[%s] msgId %d not found handler", b.name, msg.GetMsgId())
		return
	}
	f(msg)
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

	handler := NewBaseActorHandler("actor_handler")

	// 直接复制 handlerMap（高性能）
	handler.handlerMap = make(map[uint16]HandlerMessageFunc, len(b.handlerMap))
	for msgId, f := range b.handlerMap {
		handler.handlerMap[msgId] = f
	}
	handler.name = b.name
	return handler
}

func (b *BaseActorHandler) SetActorContext(_ IActorContext) {}
