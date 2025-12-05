package playeractor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/gshare"
)

var _ actor.IActorHandler = (*PlayerHandler)(nil)

// NewPlayerHandler 创建玩家消息处理器
func NewPlayerHandler() *PlayerHandler {
	return &PlayerHandler{
		BaseActorHandler: actor.NewBaseActorHandler("player handler"),
	}
}

// PlayerHandler 玩家消息处理器
type PlayerHandler struct {
	*actor.BaseActorHandler
	actorCtx actor.IActorContext // 存储 Actor Context 引用
}

// SetActorContext 设置 Actor Context 引用
func (h *PlayerHandler) SetActorContext(ctx actor.IActorContext) {
	h.actorCtx = ctx
}

func (h *PlayerHandler) Loop() {
	if h.actorCtx != nil {
		sessionId := h.actorCtx.GetData("key") // 在这使用 sessionId做的key
		ctx := context.Background()
		ctx = context.WithValue(ctx, gshare.ContextKeySession, sessionId)
		h.actorCtx.ExecuteAsync(actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdDoRunOneMsg), []byte{}))
	}
}
