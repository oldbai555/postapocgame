package fuben

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
)

// GetDungeonRecordUseCase 获取副本记录用例
// 负责根据副本ID和难度查找副本记录（纯业务逻辑）
type GetDungeonRecordUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewGetDungeonRecordUseCase 创建获取副本记录用例
func NewGetDungeonRecordUseCase(playerRepo repository.PlayerRepository) *GetDungeonRecordUseCase {
	return &GetDungeonRecordUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行获取副本记录用例
func (uc *GetDungeonRecordUseCase) Execute(ctx context.Context, roleID uint64, dungeonID uint32, difficulty uint32) (*protocol.DungeonRecord, error) {
	dungeonData, err := deps.PlayerGateway().GetDungeonData(ctx)
	if err != nil {
		return nil, err
	}

	// 查找匹配的副本记录
	for _, record := range dungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record, nil
		}
	}
	return nil, nil
}
