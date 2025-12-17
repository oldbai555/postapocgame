package equip

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/gevent"
)

// EquipItemUseCase 装备物品用例（小 service 风格，持有 Deps）
type EquipItemUseCase struct {
	deps Deps
}

// NewEquipItemUseCase 创建装备物品用例
func NewEquipItemUseCase(deps Deps) *EquipItemUseCase {
	return &EquipItemUseCase{
		deps: deps,
	}
}

// Execute 执行装备物品用例
func (uc *EquipItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, slot uint32) error {
	// 检查物品配置
	itemConfig := uc.deps.ConfigManager.GetItemConfig(itemID)
	if itemConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	// 检查是否为装备
	if itemConfig.Type != uint32(protocol.ItemType_ItemTypeEquipment) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item is not equipment")
	}

	// 检查槽位是否匹配（使用SubType，根据Type确定含义）
	// SubType对于装备类型表示EquipSlot
	if itemConfig.SubType != slot {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "equip slot mismatch")
	}

	// 检查背包中是否有该物品
	if uc.deps.BagUseCase == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag use case not initialized")
	}
	hasItem, err := uc.deps.BagUseCase.HasItem(ctx, roleID, itemID, 1)
	if err != nil {
		return err
	}
	if !hasItem {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not in bag")
	}

	equipData, err := uc.deps.PlayerRepo.GetEquipData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 检查槽位是否已有装备
	oldEquip := uc.findEquipBySlot(equipData.Equips, slot)

	// 从背包移除物品
	if err := uc.deps.BagUseCase.RemoveItem(ctx, roleID, itemID, 1); err != nil {
		return err
	}

	// 如果有旧装备，先移除
	if oldEquip != nil {
		// 从列表中移除
		for i, e := range equipData.Equips {
			if e != nil && e.Slot == slot {
				equipData.Equips = append(equipData.Equips[:i], equipData.Equips[i+1:]...)
				break
			}
		}
		// 放回背包
		if err := uc.deps.BagUseCase.AddItem(ctx, roleID, oldEquip.ItemId, 1, 1); err != nil {
			// 记录错误但不返回，避免影响装备流程
			// log.Errorf("add old equip to bag failed: %v", err)
		}
	}

	// 添加新装备
	newEquip := &protocol.EquipSt{
		ItemId:       itemID,
		Slot:         slot,
		Level:        1,
		Exp:          0,
		RefineLevel:  0,
		EnchantAttrs: make([]*protocol.AttrSt, 0),
		SetId:        0,
	}
	equipData.Equips = append(equipData.Equips, newEquip)

	// 发布事件
	uc.deps.EventPublisher.PublishPlayerEvent(ctx, gevent.OnEquipChange, map[string]interface{}{
		"slot":    slot,
		"item_id": itemID,
	})

	return nil
}

// findEquipBySlot 根据槽位查找装备
func (uc *EquipItemUseCase) findEquipBySlot(equips []*protocol.EquipSt, slot uint32) *protocol.EquipSt {
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
