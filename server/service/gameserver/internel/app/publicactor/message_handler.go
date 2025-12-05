package publicactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/servertime"
)

// handleRunOneMsg 处理 RunOne 消息（需要在 Handler 中调用）
func handleRunOneMsg(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	// 定期刷新离线数据到数据库（每 60 秒刷新一次）
	now := servertime.Now().UnixMilli()
	publicRole.flushOfflineDataIfNeeded(ctx, now)

	// 其他 RunOne 逻辑（如清理离线消息、刷新排行榜等）
	// ...
}
