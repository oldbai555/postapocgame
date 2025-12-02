package repository

import (
	"context"
	"postapocgame/server/internal/protocol"
)

// PlayerRepository 玩家数据访问接口（Domain 层定义）
type PlayerRepository interface {
	// GetBinaryData 获取玩家 BinaryData（返回共享引用，不复制）
	GetBinaryData(ctx context.Context, roleID uint64) (*protocol.PlayerRoleBinaryData, error)

	// SaveBinaryData 保存玩家 BinaryData
	SaveBinaryData(ctx context.Context, roleID uint64, binaryData *protocol.PlayerRoleBinaryData) error
}
