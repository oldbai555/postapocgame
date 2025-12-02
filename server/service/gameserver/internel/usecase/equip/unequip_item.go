package equip

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// UnEquipItemUseCase 卸载装备用例
type UnEquipItemUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
	bagUseCase     interfaces.BagUseCase // 依赖 BagSys
}

// NewUnEquipItemUseCase 创建卸载装备用例
func NewUnEquipItemUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
) *UnEquipItemUseCase {
	return &UnEquipItemUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *UnEquipItemUseCase) SetDependencies(bagUseCase interfaces.BagUseCase) {
	uc.bagUseCase = bagUseCase
}

// Execute 执行卸载装备用例
func (uc *UnEquipItemUseCase) Execute(ctx context.Context, roleID uint64, slot uint32) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	if binaryData.EquipData == nil || binaryData.EquipData.Equips == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	// 查找装备
	equip := uc.findEquipBySlot(binaryData.EquipData.Equips, slot)
	if equip == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	// 添加到背包
	if uc.bagUseCase != nil {
		if err := uc.bagUseCase.AddItem(ctx, roleID, equip.ItemId, 1, 1); err != nil {
			return err
		}
	} else {
		// 旧方式：通过 GetBagSys 获取（向后兼容）
		return uc.unEquipItemLegacy(ctx, roleID, slot)
	}

	// 从列表中移除
	for i, e := range binaryData.EquipData.Equips {
		if e != nil && e.Slot == slot {
			binaryData.EquipData.Equips = append(binaryData.EquipData.Equips[:i], binaryData.EquipData.Equips[i+1:]...)
			break
		}
	}

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnEquipChange, map[string]interface{}{
		"slot": slot,
	})

	return nil
}

// findEquipBySlot 根据槽位查找装备
func (uc *UnEquipItemUseCase) findEquipBySlot(equips []*protocol.EquipSt, slot uint32) *protocol.EquipSt {
	if equips == nil {
		return nil
	}
	for _, equip := range equips {
		if equip != nil && equip.Slot == slot {
			return equip
		}
	}
	return nil
}

// unEquipItemLegacy 卸载装备（旧方式，向后兼容）
func (uc *UnEquipItemUseCase) unEquipItemLegacy(ctx context.Context, roleID uint64, slot uint32) error {
	// 暂时不处理，等 BagSys 完全重构后通过接口调用
	_ = ctx
	_ = roleID
	_ = slot
	return nil
}
