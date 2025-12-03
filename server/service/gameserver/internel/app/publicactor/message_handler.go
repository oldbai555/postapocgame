package publicactor

import (
	"context"
	"postapocgame/server/internal/actor"
)

// handleRunOneMsg 处理 RunOne 消息（需要在 Handler 中调用）
// 注意：PublicRole 目前不需要 RunOne 方法，此函数保留用于未来扩展
func handleRunOneMsg(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	// PublicRole 目前不需要每帧执行逻辑
	// 如果需要，可以在这里添加清理离线消息、刷新排行榜等逻辑
}
