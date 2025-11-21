package publicactor

import (
	"context"
	"postapocgame/server/internal/actor"
)

// handleRunOneMsg 处理 RunOne 消息（需要在 Handler 中调用）
func handleRunOneMsg(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	if publicRole != nil {
		publicRole.RunOne(ctx)
	}
}
