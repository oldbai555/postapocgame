package iface

import (
	"context"
)

type IDungeonServer interface {
	// Start 启动服务器
	Start(ctx context.Context) error
	// Stop 停止服务器
	Stop(ctx context.Context) error
}

// IGameServerRPC GameServer RPC接口
type IGameServerRPC interface {
	// Call 调用GameServer RPC
	Call(ctx context.Context, platformId, zoneId uint32, msgId uint16, data []byte) ([]byte, error)
}
