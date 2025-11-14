package actor

import (
	"context"
	"fmt"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"sync"
)

// NewActorManager 创建Actor管理器
func NewActorManager(mode ActorMode, mailboxSize int, actorHandlerFactoryFunc func() IActorHandler) IActorManager {
	return &actorManager{
		mode:                    mode,
		mailboxSize:             mailboxSize,
		actorHandlerFactoryFunc: actorHandlerFactoryFunc,
	}
}

type actorManager struct {
	mode                    ActorMode
	actors                  sync.Map
	mailboxSize             int
	actorHandlerFactoryFunc func() IActorHandler
	ctx                     context.Context
	cancel                  context.CancelFunc
}

func (m *actorManager) GetMode() ActorMode {
	return m.mode
}

func (m *actorManager) Init() error {
	return nil
}

func (m *actorManager) Start(ctx context.Context) error {
	if m.actorHandlerFactoryFunc == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "not found msg handler factory func")
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	// 创建单例Actor
	if m.mode == ModeSingle {
		actorHandler := m.actorHandlerFactoryFunc()
		key := "global"
		actor := newActorContext(key, m.mailboxSize, WithIActorHandler(actorHandler))
		actor.start()
		actorHandler.SetActorContext(actor)
		actor.SetData("key", key)
		m.actors.Store(key, actor)
	}

	return nil
}

func (m *actorManager) Stop(ctx context.Context) error {
	m.cancel()
	// 停止所有Actor
	m.actors.Range(func(key, value any) bool {
		actor, ok := value.(*actorContext)
		if !ok {
			log.Errorf("key:%v,value:%v convert actor context failed", key, value)
			return true
		}
		actor.stop()
		return true
	})
	m.actors.Clear()
	return nil
}

func (m *actorManager) GetOrCreateActor(key string) (IActorContext, error) {
	// 总是返回全局Actor
	if m.mode == ModeSingle {
		key = "global"
	}
	actor, ok := m.getActor(key)
	if ok {
		return actor, nil
	}
	actorHandler := m.actorHandlerFactoryFunc()
	actor = newActorContext(key, m.mailboxSize, WithIActorHandler(actorHandler))
	actor.start()
	m.actors.Store(key, actor)
	actor.SetData("key", key)
	actorHandler.SetActorContext(actor)
	return actor, nil
}

func (m *actorManager) getActor(key string) (*actorContext, bool) {
	if m.mode == ModeSingle {
		key = "global"
	}

	value, exists := m.actors.Load(key)
	if exists {
		return value.(*actorContext), true
	}
	return nil, false
}

func (m *actorManager) RemoveActor(key string) error {
	// 不允许移除全局Actor
	if m.mode == ModeSingle {
		return fmt.Errorf("cannot remove global actor in single mode")
	}
	value, exists := m.actors.Load(key)
	if !exists {
		return nil
	}
	m.actors.Delete(key)
	value.(*actorContext).stop()
	return nil
}

func (m *actorManager) BroadcastAsync(message IActorMessage) {
	m.actors.Range(func(key, value any) bool {
		actor := value.(*actorContext)
		actor.ExecuteAsync(message)
		return true
	})

}
func (m *actorManager) SendMessageAsync(key string, message IActorMessage) error {
	actor, err := m.GetOrCreateActor(key)
	if err != nil {
		return customerr.Wrap(err)
	}
	actor.ExecuteAsync(message)
	return nil
}
