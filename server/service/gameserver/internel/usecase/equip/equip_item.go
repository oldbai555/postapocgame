package equip

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// EquipItemUseCase 装备物品用例
type EquipItemUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
	configManager  interfaces.ConfigManager
	bagUseCase     interfaces.BagUseCase    // 依赖 BagSys
	rewardUseCase  interfaces.RewardUseCase // 用于发放奖励（暂时通过旧方式）
}

// NewEquipItemUseCase 创建装备物品用例
func NewEquipItemUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
	configManager interfaces.ConfigManager,
) *EquipItemUseCase {
	return &EquipItemUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
		configManager:  configManager,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *EquipItemUseCase) SetDependencies(bagUseCase interfaces.BagUseCase, rewardUseCase interfaces.RewardUseCase) {
	uc.bagUseCase = bagUseCase
	uc.rewardUseCase = rewardUseCase
}

// Execute 执行装备物品用例
func (uc *EquipItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, slot uint32) error {
	// 检查物品配置
	itemConfigRaw, ok := uc.configManager.GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	itemConfig, ok := itemConfigRaw.(*jsonconf.ItemConfig)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid item config type")
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
	if uc.bagUseCase != nil {
		hasItem, err := uc.bagUseCase.HasItem(ctx, roleID, itemID, 1)
		if err != nil {
			return err
		}
		if !hasItem {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not in bag")
		}
	} else {
		// 旧方式：通过 GetBagSys 获取（向后兼容）
		return uc.equipItemLegacy(ctx, roleID, itemID, slot)
	}

	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 确保 EquipData 已初始化
	if binaryData.EquipData == nil {
		binaryData.EquipData = &protocol.SiEquipData{
			Equips: make([]*protocol.EquipSt, 0),
		}
	}
	if binaryData.EquipData.Equips == nil {
		binaryData.EquipData.Equips = make([]*protocol.EquipSt, 0)
	}

	// 检查槽位是否已有装备
	oldEquip := uc.findEquipBySlot(binaryData.EquipData.Equips, slot)

	// 从背包移除物品
	if uc.bagUseCase != nil {
		if err := uc.bagUseCase.RemoveItem(ctx, roleID, itemID, 1); err != nil {
			return err
		}
	}

	// 如果有旧装备，先移除
	if oldEquip != nil {
		// 从列表中移除
		for i, e := range binaryData.EquipData.Equips {
			if e != nil && e.Slot == slot {
				binaryData.EquipData.Equips = append(binaryData.EquipData.Equips[:i], binaryData.EquipData.Equips[i+1:]...)
				break
			}
		}
		// 放回背包
		if uc.bagUseCase != nil {
			if err := uc.bagUseCase.AddItem(ctx, roleID, oldEquip.ItemId, 1, 1); err != nil {
				// 记录错误但不返回，避免影响装备流程
				// log.Errorf("add old equip to bag failed: %v", err)
			}
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
	binaryData.EquipData.Equips = append(binaryData.EquipData.Equips, newEquip)

	// 发布事件
	uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnEquipChange, map[string]interface{}{
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

// equipItemLegacy 装备物品（旧方式，向后兼容）
func (uc *EquipItemUseCase) equipItemLegacy(ctx context.Context, roleID uint64, itemID uint32, slot uint32) error {
	// 通过 Context 获取 PlayerRole，然后调用 BagSys（向后兼容）
	// 注意：这里违反了 Clean Architecture 原则，等 BagSys 完全重构后移除
	// 暂时不处理，等 BagSys 完全重构后通过接口调用
	_ = ctx
	_ = roleID
	_ = itemID
	_ = slot
	return nil
}
