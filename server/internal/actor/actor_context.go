/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package actor

import (
	"postapocgame/server/pkg/routine"
	"sync"
	"sync/atomic"
)

type actorContext struct {
	id       string
	running  atomic.Bool
	handler  IActorHandler
	mailbox  chan IActorMessage
	stopChan chan struct{}
	wg       sync.WaitGroup

	data interface{}
}

func newActorContext(id string, mailboxSize int, opts ...ContextOption) *actorContext {
	a := &actorContext{
		id:       id,
		mailbox:  make(chan IActorMessage, mailboxSize),
		stopChan: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *actorContext) GetID() string {
	return a.id
}

func (a *actorContext) ExecuteAsync(message IActorMessage) {
	if !a.IsRunning() {
		return
	}

	select {
	case a.mailbox <- message:
	case <-a.stopChan:
		return
	default:
		// 邮箱满了，丢弃消息
		// 业务层可以通过日志等方式处理
	}
}

func (a *actorContext) GetData() interface{} {
	return a.data
}

func (a *actorContext) SetData(data interface{}) {
	a.data = data
}

func (a *actorContext) IsRunning() bool {
	return a.running.Load()
}

func (a *actorContext) start() {
	a.running.Store(true)
	if a.handler != nil {
		a.handler.OnStart()
	}
	a.wg.Add(1)
	routine.GoV2(func() error {
		defer a.wg.Done()
		a.loop()
		return nil
	})
}

func (a *actorContext) stop() {
	if !a.running.Load() {
		return
	}
	a.running.Store(false)
	if a.handler != nil {
		a.handler.OnStop()
	}
	close(a.stopChan)
	a.wg.Wait()
}

func (a *actorContext) loop() {
	var doMsgLogic = func(msg IActorMessage) {
		if msg == nil {
			return
		}
		if a.handler == nil {
			return
		}
		routine.Run(func() {
			a.handler.HandleMessage(msg)
		})
	}
	for {
		if a.handler != nil {
			routine.Run(func() {
				a.handler.Loop()
			})
		}
		select {
		case msg := <-a.mailbox:
			doMsgLogic(msg)
		case <-a.stopChan:
			// 处理剩余消息
			for {
				select {
				case msg := <-a.mailbox:
					doMsgLogic(msg)
				default:
					return
				}
			}
		}
	}
}

type ContextOption func(actorCtx *actorContext)

func WithIActorHandler(handler IActorHandler) ContextOption {
	return func(actorCtx *actorContext) {
		actorCtx.handler = handler
	}
}
