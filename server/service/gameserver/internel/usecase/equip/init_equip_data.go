package equip

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// InitEquipDataUseCase 初始化装备数据用例
// 负责装备数据的初始化（装备列表结构）
type InitEquipDataUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewInitEquipDataUseCase 创建初始化装备数据用例
func NewInitEquipDataUseCase(
	playerRepo repository.PlayerRepository,
) *InitEquipDataUseCase {
	return &InitEquipDataUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行初始化装备数据用例
func (uc *InitEquipDataUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 如果equip_data不存在，则初始化
	if binaryData.EquipData == nil {
		binaryData.EquipData = &protocol.SiEquipData{
			Equips: make([]*protocol.EquipSt, 0),
		}
	}

	return nil
}
