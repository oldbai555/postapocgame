package gateway

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/gshare"
)

var _ repository.PlayerRepository = (*PlayerGateway)(nil)

// PlayerGateway 玩家数据访问实现（实现 domain 层的 Repository 接口）
type PlayerGateway struct{}

// NewPlayerGateway 创建玩家 Gateway
func NewPlayerGateway() repository.PlayerRepository {
	return &PlayerGateway{}
}

func (g *PlayerGateway) GetBagData(ctx context.Context) (*protocol.SiBagData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrBagDataNotFound
	}
	bagData := playerRole.GetBagData()
	if bagData == nil {
		return nil, repository.ErrBagDataNotFound
	}
	return bagData, nil
}

func (g *PlayerGateway) GetMoneyData(ctx context.Context) (*protocol.SiMoneyData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrMoneyDataNotFound
	}
	moneyData := playerRole.GetMoneyData()
	if moneyData == nil {
		return nil, repository.ErrMoneyDataNotFound
	}
	return moneyData, nil
}

func (g *PlayerGateway) GetLevelData(ctx context.Context) (*protocol.SiLevelData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrLevelDataNotFound
	}
	levelData := playerRole.GetLevelData()
	if levelData == nil {
		return nil, repository.ErrLevelDataNotFound
	}
	return levelData, nil
}

func (g *PlayerGateway) GetEquipData(ctx context.Context) (*protocol.SiEquipData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrEquipDataNotFound
	}
	equipData := playerRole.GetEquipData()
	if equipData == nil {
		return nil, repository.ErrEquipDataNotFound
	}
	return equipData, nil
}

func (g *PlayerGateway) GetSkillData(ctx context.Context) (*protocol.SiSkillData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrSkillDataNotFound
	}
	skillData := playerRole.GetSkillData()
	if skillData == nil {
		return nil, repository.ErrSkillDataNotFound
	}
	return skillData, nil
}

func (g *PlayerGateway) GetItemUseData(ctx context.Context) (*protocol.SiItemUseData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrItemUseDataNotFound
	}
	itemUseData := playerRole.GetItemUseData()
	if itemUseData == nil {
		return nil, repository.ErrItemUseDataNotFound
	}
	return itemUseData, nil
}

func (g *PlayerGateway) GetQuestData(ctx context.Context) (*protocol.SiQuestData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrQuestDataNotFound
	}
	questData := playerRole.GetQuestData()
	if questData == nil {
		return nil, repository.ErrQuestDataNotFound
	}
	return questData, nil
}

func (g *PlayerGateway) GetDungeonData(ctx context.Context) (*protocol.SiDungeonData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrDungeonDataNotFound
	}
	dungeonData := playerRole.GetDungeonData()
	if dungeonData == nil {
		return nil, repository.ErrDungeonDataNotFound
	}
	return dungeonData, nil
}

func (g *PlayerGateway) GetMailData(ctx context.Context) (*protocol.SiMailData, error) {
	// 优先从 Context 中的 PlayerRole 获取（共享引用）
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil {
		return nil, repository.ErrMailDataNotFound
	}
	mailData := playerRole.GetMailData()
	if mailData == nil {
		return nil, repository.ErrMailDataNotFound
	}
	return mailData, nil
}
