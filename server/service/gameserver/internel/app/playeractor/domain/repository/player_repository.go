package repository

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

var (
	// ErrBagDataNotFound 背包数据不存在
	ErrBagDataNotFound = customerr.NewError("bag data not found")
	// ErrMoneyDataNotFound 货币数据不存在
	ErrMoneyDataNotFound = customerr.NewError("money data not found")
	// ErrLevelDataNotFound 等级数据不存在
	ErrLevelDataNotFound = customerr.NewError("level data not found")
	// ErrEquipDataNotFound 装备数据不存在
	ErrEquipDataNotFound = customerr.NewError("equip data not found")
	// ErrSkillDataNotFound 技能数据不存在
	ErrSkillDataNotFound = customerr.NewError("skill data not found")
	// ErrItemUseDataNotFound 物品使用数据不存在
	ErrItemUseDataNotFound = customerr.NewError("item use data not found")
	// ErrQuestDataNotFound 任务数据不存在
	ErrQuestDataNotFound = customerr.NewError("quest data not found")
	// ErrDungeonDataNotFound 副本数据不存在
	ErrDungeonDataNotFound = customerr.NewError("dungeon data not found")
	// ErrMailDataNotFound 邮件数据不存在
	ErrMailDataNotFound = customerr.NewError("mail data not found")
)

// PlayerRepository 玩家数据访问接口（Domain 层定义）
type PlayerRepository interface {
	GetBagData(ctx context.Context) (*protocol.SiBagData, error)
	GetMoneyData(ctx context.Context) (*protocol.SiMoneyData, error)
	GetLevelData(ctx context.Context) (*protocol.SiLevelData, error)
	GetEquipData(ctx context.Context) (*protocol.SiEquipData, error)
	GetSkillData(ctx context.Context) (*protocol.SiSkillData, error)
	GetItemUseData(ctx context.Context) (*protocol.SiItemUseData, error)
	GetQuestData(ctx context.Context) (*protocol.SiQuestData, error)
	GetDungeonData(ctx context.Context) (*protocol.SiDungeonData, error)
	GetMailData(ctx context.Context) (*protocol.SiMailData, error)
}
