package dungeonactor

import (
	"sync"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/fbmgr"
)

var _ actor.IActorHandler = (*DungeonActorHandler)(nil)

// DungeonActorHandler 处理 DungeonActor 的消息
type DungeonActorHandler struct {
	*actor.BaseActorHandler
	inLoop sync.Mutex
}

// NewDungeonActorHandler 创建 DungeonActor 消息处理器
func NewDungeonActorHandler() *DungeonActorHandler {
	h := &DungeonActorHandler{
		BaseActorHandler: actor.NewBaseActorHandler("dungeon_actor_handler"),
	}
	h.OnInit()

	// 初始化默认副本
	if err := fbmgr.GetFuBenMgr().CreateDefaultFuBen(); err != nil {
		log.Errorf("[dungeonactor] failed to create default fuben: %v", err)
	}

	return h
}

// Loop Actor 单线程循环，驱动实体和副本管理
func (h *DungeonActorHandler) Loop() {
	if h == nil {
		return
	}
	h.inLoop.Lock()
	defer h.inLoop.Unlock()

	now := servertime.Now()
	entitymgr.RunOne(now)
	fbmgr.GetFuBenMgr().RunOne(now)

	log.Debugf("[dungeonactor] tick at %v (RunOne entitymgr/fbmgr)", now)
}

// HandleMessage 处理 Actor 消息
func (h *DungeonActorHandler) HandleMessage(msg actor.IActorMessage) {
	h.BaseActorHandler.HandleMessage(msg)
}
