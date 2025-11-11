package playeractor

import (
	"postapocgame/server/internal/actor"
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
	return
}

func (h *PlayerHandler) OnStart() {

}

func (h *PlayerHandler) OnStop() {

}
