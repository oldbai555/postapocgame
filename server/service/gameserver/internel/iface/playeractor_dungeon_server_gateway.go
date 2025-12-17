package iface

import "context"

// DungeonServerGateway DungeonServer Gateway 接口（Use Case 层定义）
// 用于 PlayerActor 向 DungeonActor 发送消息
type DungeonServerGateway interface {
	// AsyncCall 异步调用 DungeonActor（DungeonActor 为单例，不需要 srvType）
	AsyncCall(ctx context.Context, sessionId string, msgId uint16, data []byte) error
}
