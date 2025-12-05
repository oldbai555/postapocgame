package publicactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/routine"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// PublicActor 公共Actor适配器
type PublicActor struct {
	actorMgr      actor.IActorManager
	mode          actor.ActorMode
	publicHandler *PublicHandler
}

func NewPublicActor() *PublicActor {
	defaultHandler := NewPublicHandler()
	defaultHandler.BaseActorHandler = actor.NewBaseActorHandler("public actor handler")
	defaultHandler.OnInit()
	p := &PublicActor{
		mode:          actor.ModeSingle,
		publicHandler: defaultHandler,
	}
	p.actorMgr = actor.NewActorManager(actor.ModeSingle, 1000, func() actor.IActorHandler {
		return defaultHandler
	})

	// 先注册到 gshare，然后再注册消息处理器
	gshare.SetPublicActorFacade(p)

	return p
}

func (p *PublicActor) RegisterHandler(msgId uint16, f actor.HandlerMessageFunc) {
	routine.Run(func() {
		p.publicHandler.RegisterMessageHandler(msgId, f)
	})
}

func (p *PublicActor) SendMessageAsync(key string, message actor.IActorMessage) error {
	var err error
	routine.Run(func() {
		// PublicActor 是单例，key 固定为 "global"
		err = p.actorMgr.SendMessageAsync("global", message)
	})
	return err
}

func (p *PublicActor) Init() error {
	var err error
	routine.Run(func() {
		err = p.actorMgr.Init()
		if err == nil {
			p.publicHandler.OnInit()
		}
	})
	if err != nil {
		return customerr.Wrap(err)
	}
	return nil
}

// Start 启动Actor系统
func (p *PublicActor) Start(ctx context.Context) error {
	var err error
	routine.Run(func() {
		err = p.actorMgr.Start(ctx)
	})
	return err
}

// Stop 停止Actor系统
func (p *PublicActor) Stop(ctx context.Context) error {
	var err error
	routine.Run(func() {
		err = p.actorMgr.Stop(ctx)
	})
	return err
}
