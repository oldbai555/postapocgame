/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package dungeonactor

import (
	"postapocgame/server/internal/actor"
)

type ActorMsgHandler struct {
	actor.BaseActorMsgHandler
}

func (h *ActorMsgHandler) OnInit() {
	h.BaseActorMsgHandler.OnInit()
}
