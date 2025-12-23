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
	defaultHandler.BaseActorHandler = actor.NewBaseActorHandler("player role handler")
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
	// 同步注册消息处理器，避免协程内对 handler 的并发访问
	p.playerHandler.RegisterMessageHandler(msgId, f)
}

func (p *PlayerRoleActor) SendMessageAsync(key string, message actor.IActorMessage) error {
	// 直接转发给底层 ActorManager，由调用方决定是否需要在更外层异步化
	return p.actorMgr.SendMessageAsync(key, message)
}

func (p *PlayerRoleActor) RemoveActor(key string) error {
	return p.actorMgr.RemoveActor(key)
}

func (p *PlayerRoleActor) Init() error {
	if err := p.actorMgr.Init(); err != nil {
		return customerr.Wrap(err)
	}
	// 在 ActorManager 初始化完成后，再初始化默认的 handler 模板
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
