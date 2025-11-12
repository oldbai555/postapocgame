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
	rw         sync.RWMutex
	handlerMap map[uint16]HandlerMessageFunc
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
}

func (b *BaseActorHandler) HandleMessage(msg IActorMessage) {
	b.rw.RLock()
	f := b.handlerMap[msg.GetMsgId()]
	b.rw.RUnlock() // ✅ 修复：RLock 后应该 RUnlock

	if f == nil {
		log.Errorf("msgId %d not found handler", msg.GetMsgId())
		return
	}
	routine.Run(func() {
		f(msg)
	})
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

	// 直接复制 handlerMap（高性能）
	handler.handlerMap = make(map[uint16]HandlerMessageFunc, len(b.handlerMap))
	for msgId, f := range b.handlerMap {
		handler.handlerMap[msgId] = f
	}
	return handler
}
