package dungeonactor

import (
	"postapocgame/server/service/gameserver/internel/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/dungeonactor/fbmgr"
	"sync"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
)

var _ actor.IActorHandler = (*Handler)(nil)

type Handler struct {
	*actor.BaseActorHandler
	inLoop sync.Mutex
}

// NewDungeonActorHandler 创建 DungeonActor 消息处理器
func NewDungeonActorHandler() *Handler {
	h := &Handler{
		BaseActorHandler: actor.NewBaseActorHandler("dungeon_actor_handler"),
	}
	h.OnInit()

	// 初始化默认副本
	if err := fbmgr.GetFuBenMgr().CreateDefaultFuBen(); err != nil {
		log.Errorf("[dungeon-actor] failed to create default fuben: %v", err)
	}

	return h
}

// Loop Actor 单线程循环，驱动实体和副本管理
func (h *Handler) Loop() {
	if h == nil {
		return
	}
	h.inLoop.Lock()
	defer h.inLoop.Unlock()

	now := servertime.Now()
	entitymgr.RunOne(now)
	fbmgr.GetFuBenMgr().RunOne(now)
}

// HandleMessage 处理 Actor 消息
func (h *Handler) HandleMessage(msg actor.IActorMessage) {
	h.BaseActorHandler.HandleMessage(msg)
}
