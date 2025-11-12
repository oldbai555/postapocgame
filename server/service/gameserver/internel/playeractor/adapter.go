package playeractor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// PlayerRoleActor 玩家角色Actor适配器
type PlayerRoleActor struct {
	actorMgr      actor.IActorManager
	mode          actor.ActorMode
	playerHandler *PlayerHandler
}

func NewPlayerRoleActor(mode actor.ActorMode) *PlayerRoleActor {
	defaultHandler := NewPlayerHandler()
	defaultHandler.BaseActorHandler = actor.NewBaseActorHandler()
	defaultHandler.OnInit()
	p := &PlayerRoleActor{
		mode:          mode,
		playerHandler: defaultHandler,
	}
	p.actorMgr = actor.NewActorManager(mode, 1000, p.NewPlayerHandlerFactory)

	// 🔧 使用接口方式注册
	gshare.SetActorFacade(p)

	return p
}

func (p *PlayerRoleActor) RegisterHandler(msgId uint16, f actor.HandlerMessageFunc) {
	p.playerHandler.RegisterMessageHandler(msgId, f)
}

func (p *PlayerRoleActor) SendMessageAsync(key string, message actor.IActorMessage) error {
	return p.actorMgr.SendMessageAsync(key, message)
}

func (p *PlayerRoleActor) RemoveActor(key string) error {
	return p.actorMgr.RemoveActor(key)
}

func (p *PlayerRoleActor) Init() error {
	err := p.actorMgr.Init()
	if err != nil {
		return customerr.Wrap(err)
	}
	p.playerHandler.OnInit()
	return nil
}

// Start 启动Actor系统
func (p *PlayerRoleActor) Start(ctx context.Context) error {
	return p.actorMgr.Start(ctx)
}

// Stop 停止Actor系统
func (p *PlayerRoleActor) Stop(ctx context.Context) error {
	return p.actorMgr.Stop(ctx)
}

func (p *PlayerRoleActor) NewPlayerHandlerFactory() actor.IActorHandler {
	playerHandler := NewPlayerHandler()
	playerHandler.BaseActorHandler = p.playerHandler.Clone()
	return playerHandler
}
