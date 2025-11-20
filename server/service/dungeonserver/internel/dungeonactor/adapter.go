/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package dungeonactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/service/dungeonserver/internel/dshare"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/fbmgr"
	"sync/atomic"
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

type dungeonActorHandler struct {
	*actor.BaseActorHandler
	inLoop atomic.Bool
}

func newDungeonActorHandler() *dungeonActorHandler {
	return &dungeonActorHandler{
		BaseActorHandler: actor.NewBaseActorHandler(),
	}
}

func (h *dungeonActorHandler) Loop() {
	if h == nil {
		return
	}
	if !h.inLoop.CompareAndSwap(false, true) {
		return
	}
	defer h.inLoop.Store(false)
	now := servertime.Now()
	entitymgr.RunOne(now)
	fbmgr.GetFuBenMgr().RunOne(now)
}

func NewDungeonActor(mode actor.ActorMode) *DungeonActor {
	handler := newDungeonActorHandler()
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
