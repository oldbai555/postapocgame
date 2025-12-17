package fuben

import (
	"context"
	"postapocgame/server/internal/protocol"
)

// GetDungeonRecordUseCase 获取副本记录用例（小 service 风格，持有 Deps）
type GetDungeonRecordUseCase struct {
	deps Deps
}

// NewGetDungeonRecordUseCase 创建获取副本记录用例
func NewGetDungeonRecordUseCase(deps Deps) *GetDungeonRecordUseCase {
	return &GetDungeonRecordUseCase{deps: deps}
}

// Execute 执行获取副本记录用例
func (uc *GetDungeonRecordUseCase) Execute(ctx context.Context, roleID uint64, dungeonID uint32, difficulty uint32) (*protocol.DungeonRecord, error) {
	dungeonData, err := uc.deps.PlayerRepo.GetDungeonData(ctx)
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
