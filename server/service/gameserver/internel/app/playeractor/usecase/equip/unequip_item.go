package equip

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// UnEquipItemUseCase 卸载装备用例
type UnEquipItemUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces2.EventPublisher
	bagUseCase     interfaces2.BagUseCase // 依赖 BagSys
}

// NewUnEquipItemUseCase 创建卸载装备用例
func NewUnEquipItemUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces2.EventPublisher,
) *UnEquipItemUseCase {
	return &UnEquipItemUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *UnEquipItemUseCase) SetDependencies(bagUseCase interfaces2.BagUseCase) {
	uc.bagUseCase = bagUseCase
}

// Execute 执行卸载装备用例
func (uc *UnEquipItemUseCase) Execute(ctx context.Context, roleID uint64, slot uint32) error {
	equipData, err := deps.PlayerGateway().GetEquipData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 查找装备
	equip := uc.findEquipBySlot(equipData.Equips, slot)
	if equip == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	// 添加到背包
	if uc.bagUseCase == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag use case not initialized")
	}
	if err := uc.bagUseCase.AddItem(ctx, roleID, equip.ItemId, 1, 1); err != nil {
		return err
	}

	// 从列表中移除
	for i, e := range equipData.Equips {
		if e != nil && e.Slot == slot {
			equipData.Equips = append(equipData.Equips[:i], equipData.Equips[i+1:]...)
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
