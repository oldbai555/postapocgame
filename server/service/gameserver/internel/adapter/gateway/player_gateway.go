package gateway

import (
	"context"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// PlayerGateway 玩家数据访问实现（实现 domain 层的 Repository 接口）
type PlayerGateway struct{}

// NewPlayerGateway 创建玩家 Gateway
func NewPlayerGateway() repository.PlayerRepository {
	return &PlayerGateway{}
}

// GetBinaryData 获取玩家 BinaryData（返回共享引用，不复制）
func (g *PlayerGateway) GetBinaryData(ctx context.Context, roleID uint64) (*protocol.PlayerRoleBinaryData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)
	if playerRole != nil {
		// 返回共享引用，不复制
		return playerRole.GetBinaryData(), nil
	}

	// 如果 Context 中没有，则从数据库加载（这种情况应该很少见）
	log.Warnf("PlayerRole not found in context, loading from database: roleID=%d", roleID)
	return database.GetPlayerBinaryData(uint(roleID))
}

// SaveBinaryData 保存玩家 BinaryData
func (g *PlayerGateway) SaveBinaryData(ctx context.Context, roleID uint64, binaryData *protocol.PlayerRoleBinaryData) error {
	// 保存到数据库
	return database.SavePlayerBinaryData(uint(roleID), binaryData)
}
