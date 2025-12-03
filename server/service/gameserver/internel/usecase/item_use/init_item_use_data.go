package item_use

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// InitItemUseDataUseCase 初始化物品使用数据用例
// 负责物品使用数据的初始化（冷却映射结构）
type InitItemUseDataUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewInitItemUseDataUseCase 创建初始化物品使用数据用例
func NewInitItemUseDataUseCase(
	playerRepo repository.PlayerRepository,
) *InitItemUseDataUseCase {
	return &InitItemUseDataUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行初始化物品使用数据用例
func (uc *InitItemUseDataUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 如果item_use_data不存在，则初始化
	if binaryData.ItemUseData == nil {
		binaryData.ItemUseData = &protocol.SiItemUseData{
			CooldownMap: make(map[uint32]int64),
		}
	}

	return nil
}
