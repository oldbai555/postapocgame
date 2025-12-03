package fuben

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// InitDungeonDataUseCase 初始化副本数据用例
// 负责副本数据的初始化（副本记录容器结构）
type InitDungeonDataUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewInitDungeonDataUseCase 创建初始化副本数据用例
func NewInitDungeonDataUseCase(playerRepo repository.PlayerRepository) *InitDungeonDataUseCase {
	return &InitDungeonDataUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行初始化副本数据用例
func (uc *InitDungeonDataUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 如果dungeon_data不存在，则初始化
	if binaryData.DungeonData == nil {
		binaryData.DungeonData = &protocol.SiDungeonData{
			Records: make([]*protocol.DungeonRecord, 0),
		}
	}

	// 如果Records为空，初始化为空切片
	if binaryData.DungeonData.Records == nil {
		binaryData.DungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	return nil
}
