package dungeonactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// DungeonActor GameServer 进程内的战斗/副本 Actor（单例）
// 负责处理所有副本和战斗相关的逻辑
type DungeonActor struct {
	actorMgr actor.IActorManager
	mode     actor.ActorMode
	handler  *Handler
}

// 全局唯一 DungeonActor 实例指针
var defaultDungeonActor *DungeonActor

// dungeonActorFacadeImpl 实现 gshare.IDungeonActorFacade，封装 DungeonActor 的 ActorManager。
type dungeonActorFacadeImpl struct {
	actorMgr actor.IActorManager
	handler  *Handler
}

func (f *dungeonActorFacadeImpl) RegisterHandler(msgId uint16, h actor.HandlerMessageFunc) {
	if f.handler != nil {
		f.handler.RegisterMessageHandler(msgId, h)
	}
}

func (f *dungeonActorFacadeImpl) SendMessageAsync(key string, message actor.IActorMessage) error {
	if f == nil || f.actorMgr == nil {
		return nil
	}
	return f.actorMgr.SendMessageAsync(key, message)
}

// NewDungeonActor 创建并注册全局 DungeonActor 单例
func NewDungeonActor(mode actor.ActorMode) *DungeonActor {
	if mode != actor.ModeSingle {
		log.Fatalf("[dungeon-actor] only ModeSingle is supported, got mode=%d", mode)
	}
	handler := NewDungeonActorHandler()
	d := &DungeonActor{
		mode:    mode,
		handler: handler,
	}

	d.actorMgr = actor.NewActorManager(
		mode,
		1024,
		func() actor.IActorHandler {
			return handler
		},
	)

	// 注册 DungeonActor 门面，供 PlayerActor 通过 gshare 发送内部 Actor 消息
	facade := &dungeonActorFacadeImpl{
		actorMgr: d.actorMgr,
		handler:  handler,
	}
	gshare.SetDungeonActorFacade(facade)

	defaultDungeonActor = d
	log.Infof("[dungeonactor] NewDungeonActor created with mode=%d", mode)
	return d
}

// GetDungeonActor 获取全局 DungeonActor 实例
func GetDungeonActor() *DungeonActor {
	return defaultDungeonActor
}

func (d *DungeonActor) Start(ctx context.Context) error {
	log.Infof("[dungeon-actor] Start DungeonActor")
	return d.actorMgr.Start(ctx)
}

// Stop 停止 DungeonActor
func (d *DungeonActor) Stop(ctx context.Context) error {
	log.Infof("[dungeon-actor] Stop DungeonActor")
	return d.actorMgr.Stop(ctx)
}

// AsyncCall 将消息封装为 Actor 消息，投递到 DungeonActor 的单线程 Loop 中处理
func (d *DungeonActor) AsyncCall(ctx context.Context, sessionId string, msgId uint16, data []byte) error {
	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionId)
	actorMsg := actor.NewBaseMessage(ctxWithSession, msgId, data)
	if err := d.actorMgr.SendMessageAsync("global", actorMsg); err != nil {
		log.Errorf("[dungeon-actor] AsyncCall send failed: sessionId=%s msgId=%d err=%v", sessionId, msgId, err)
		return err
	}
	return nil
}
