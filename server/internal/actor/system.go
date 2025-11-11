package actor

import (
	"context"
	"fmt"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"sync"
)

// System Actor系统实现
type System struct {
	mode Mode

	// 玩家消息处理Actor
	playerHandler   IActorMsgHandler
	singleActor     IActor            // 单Actor模式
	perPlayerActors map[string]IActor // 每玩家Actor模式

	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// NewActorSystem 创建Actor系统
func NewActorSystem(mode Mode, handler IActorMsgHandler) *System {
	return &System{
		mode:            mode,
		perPlayerActors: make(map[string]IActor),
		stopChan:        make(chan struct{}),
		playerHandler:   handler,
	}
}

func (as *System) Init() {
	if as.playerHandler != nil {
		as.playerHandler.OnInit()
	}
}

func (as *System) Start(ctx context.Context) error {
	if as.playerHandler == nil {
		return customerr.NewCustomErr("player handler is nil")
	}
	// 启动玩家消息处理Actor
	switch as.mode {
	case ModeBySingle:
		if err := as.startSingleActor(ctx); err != nil {
			return err
		}
	case ModeByPerPlayer:
		log.Infof("ActorSystem started: PerPlayer mode")
	default:
		return fmt.Errorf("unknown actor mode: %d", as.mode)
	}
	return nil
}

// Stop 停止Actor系统
func (as *System) Stop(ctx context.Context) error {
	close(as.stopChan)

	// 停止玩家消息处理Actor
	switch as.mode {
	case ModeBySingle:
		if as.singleActor != nil {
			as.singleActor.Stop(ctx)
		}
	case ModeByPerPlayer:
		as.mu.Lock()
		for sessionId, actor := range as.perPlayerActors {
			err := actor.Stop(ctx)
			if err != nil {
				log.Errorf("session:%s stop err:%v", sessionId, err)
			}
		}
		as.mu.Unlock()
	}
	as.wg.Wait()
	log.Infof("ActorSystem stopped")
	return nil
}

// Send 发送消息到Actor系统(玩家消息)
func (as *System) Send(sessionId string, msgId uint16, data []byte) error {
	msg := &Message{
		SessionId: sessionId,
		MsgId:     msgId,
		Data:      data,
		Context:   context.Background(),
	}

	switch as.mode {
	case ModeBySingle:
		return as.sendToSingleActor(msg)
	case ModeByPerPlayer:
		return as.sendToPlayerActor(msg)
	default:
		return fmt.Errorf("unknown actor mode: %d", as.mode)
	}
}

// startSingleActor 启动单Actor
func (as *System) startSingleActor(ctx context.Context) error {
	as.singleActor = NewActor(ModeBySingle, as.playerHandler)
	if err := as.singleActor.Start(ctx); err != nil {
		return err
	}
	log.Infof("ActorSystem started: Single mode")
	return nil
}

func (as *System) sendToSingleActor(msg *Message) error {
	if as.singleActor == nil {
		return fmt.Errorf("single actor not started")
	}
	return as.singleActor.Send(msg)
}

func (as *System) sendToPlayerActor(msg *Message) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	actor, ok := as.perPlayerActors[msg.SessionId]
	if !ok {
		// 创建新的玩家Actor
		actor = NewActor(ModeByPerPlayer, as.playerHandler)
		if err := actor.Start(context.Background()); err != nil {
			return err
		}
		as.perPlayerActors[msg.SessionId] = actor
		log.Infof("created player actor: %s", msg.SessionId)
	}

	return actor.Send(msg)
}

func (as *System) RemovePerPlayerActor(sessionId string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	actor, ok := as.perPlayerActors[sessionId]
	if !ok {
		return nil
	}

	err := actor.Stop(context.Background())
	if err != nil {
		log.Infof("stop player actor:%s ,err:%v", sessionId, err)
	}
	delete(as.perPlayerActors, sessionId)
	log.Infof("removed player actor: %s", sessionId)
	return nil
}
