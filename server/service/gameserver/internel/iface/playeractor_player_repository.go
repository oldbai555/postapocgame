package iface

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

var (
	// ErrLevelDataNotFound 等级数据不存在
	ErrLevelDataNotFound = customerr.NewError("level data not found")
	// ErrSkillDataNotFound 技能数据不存在
	ErrSkillDataNotFound = customerr.NewError("skill data not found")
)

// PlayerRepository 玩家数据访问接口（Domain 层定义）
type PlayerRepository interface {
	GetLevelData(ctx context.Context) (*protocol.SiLevelData, error)
	GetSkillData(ctx context.Context) (*protocol.SiSkillData, error)
}
