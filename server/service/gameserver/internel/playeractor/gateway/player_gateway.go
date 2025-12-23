package gateway

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

var _ iface.PlayerRepository = (*PlayerGateway)(nil)

// PlayerGateway 玩家数据访问实现（实现 domain 层的 Repository 接口）
type PlayerGateway struct{}

// NewPlayerGateway 创建玩家 Gateway
func NewPlayerGateway() iface.PlayerRepository {
	return &PlayerGateway{}
}

func (g *PlayerGateway) GetLevelData(ctx context.Context) (*protocol.SiLevelData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, iface.ErrLevelDataNotFound
	}
	levelData := playerRole.GetLevelData()
	if levelData == nil {
		return nil, iface.ErrLevelDataNotFound
	}
	return levelData, nil
}

func (g *PlayerGateway) GetSkillData(ctx context.Context) (*protocol.SiSkillData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, iface.ErrSkillDataNotFound
	}
	skillData := playerRole.GetSkillData()
	if skillData == nil {
		return nil, iface.ErrSkillDataNotFound
	}
	return skillData, nil
}
