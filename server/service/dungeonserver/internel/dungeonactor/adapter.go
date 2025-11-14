/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package dungeonactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/service/dungeonserver/internel/dshare"
)

type DungeonActor struct {
	actorMgr         actor.IActorManager
	mode             actor.ActorMode
	baseActorHandler actor.IActorHandler
}

func (a *DungeonActor) Start(ctx context.Context) error {
	return a.actorMgr.Start(ctx)
}

func (a *DungeonActor) Stop(ctx context.Context) error {
	return a.actorMgr.Stop(ctx)
}

func NewDungeonActor(mode actor.ActorMode) *DungeonActor {
	handler := actor.NewBaseActorHandler()
	d := &DungeonActor{
		actorMgr: actor.NewActorManager(mode, 1000, func() actor.IActorHandler {
			return handler
		}),
		mode:             mode,
		baseActorHandler: handler,
	}
	dshare.RegisterHandler = handler.RegisterMessageHandler
	dshare.SendMessageAsync = d.actorMgr.SendMessageAsync
	return d
}
