package playeractor

import (
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/routine"
)

var _ actor.IActorHandler = (*PlayerHandler)(nil)

// NewPlayerHandler 创建玩家消息处理器
func NewPlayerHandler() *PlayerHandler {
	return &PlayerHandler{}
}

// PlayerHandler 玩家消息处理器
type PlayerHandler struct {
	*actor.BaseActorHandler
}

func (h *PlayerHandler) Loop() {
	routine.Run(func() {
		// ...原本Loop的业务逻辑...
	})
}

func (h *PlayerHandler) OnStart() {
	routine.Run(func() {
		// ...业务逻辑...
	})
}

func (h *PlayerHandler) OnStop() {
	routine.Run(func() {
		// ...业务逻辑...
	})
}
