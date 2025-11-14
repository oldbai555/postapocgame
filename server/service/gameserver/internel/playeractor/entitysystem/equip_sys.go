package entitysystem

import (
	"context"
	"math"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

const (
	// 万分比基数（10000 = 100%）
	PerTenThousandBase int64 = 10000
	// 强化等级每级增加的倍率（万分比：1000 = 10%）
	UpgradePerLevelMultiplier int64 = 1000
	// 精炼等级每级增加的倍率（万分比：500 = 5%）
	RefinePerLevelMultiplier int64 = 500
	// 星级每星增加的倍率（万分比：1000 = 10%）
	StarPerLevelMultiplier int64 = 1000
	// 品质每级增加的倍率（万分比：1000 = 10%）
	QualityPerLevelMultiplier int64 = 1000
	// 阶级每级增加的倍率（万分比：1000 = 10%）
	TierPerLevelMultiplier int64 = 1000
)

// EquipSys 装备系统
type EquipSys struct {
	*BaseSystem
	equipData *protocol.SiEquipData
}

// NewEquipSys 创建装备系统
func NewEquipSys() *EquipSys {
	return &EquipSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysEquip)),
	}
}

func GetEquipSys(ctx context.Context) *EquipSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysEquip))
	if system == nil {
		log.Errorf("not found system [%v] error:%v", protocol.SystemId_SysEquip, err)
		return nil
	}
	sys := system.(*EquipSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error:%v", protocol.SystemId_SysEquip, err)
		return nil
	}
	return sys
}

// OnInit 初始化时从PlayerRoleBinaryData加载装备数据
func (es *EquipSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}

	playerID := uint(playerRole.GetPlayerRoleId())

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果equip_data不存在，则初始化
	if binaryData.EquipData == nil {
		binaryData.EquipData = &protocol.SiEquipData{
			Equips: make([]*protocol.EquipSt, 0),
		}
	}
	es.equipData = binaryData.EquipData

	log.Infof("EquipSys initialized: PlayerID=%d, EquipCount=%d", playerID, len(es.equipData.Equips))
}

// findEquipBySlot 根据槽位查找装备
func (es *EquipSys) findEquipBySlot(slot uint32) *protocol.EquipSt {
	if es.equipData == nil || es.equipData.Equips == nil {
		return nil
	}
	for _, equip := range es.equipData.Equips {
		if equip != nil && equip.Slot == slot {
			return equip
		}
	}
	return nil
}

// EquipItem 装备物品
func (es *EquipSys) EquipItem(ctx context.Context, itemID uint32, slot uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	playerID := uint(playerRole.GetPlayerRoleId())

	// 检查物品配置
	itemConfig, ok := jsonconf.GetConfigManager().GetItemConfig(itemID)
	if !ok {
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
	bagSys := GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag system not found")
	}

	item := bagSys.GetItem(itemID)
	if item == nil || item.Count == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not in bag")
	}

	// 确保equipData已初始化
	if es.equipData == nil {
		binaryData := playerRole.GetBinaryData()
		if binaryData.EquipData == nil {
			binaryData.EquipData = &protocol.SiEquipData{
				Equips: make([]*protocol.EquipSt, 0),
			}
		}
		es.equipData = binaryData.EquipData
	}
	if es.equipData.Equips == nil {
		es.equipData.Equips = make([]*protocol.EquipSt, 0)
	}

	// 检查槽位是否已有装备
	oldEquip := es.findEquipBySlot(slot)

	// 从背包移除物品
	err = bagSys.RemoveItem(ctx, itemID, 1)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 如果有旧装备，先移除
	if oldEquip != nil {
		// 从列表中移除
		for i, e := range es.equipData.Equips {
			if e != nil && e.Slot == slot {
				es.equipData.Equips = append(es.equipData.Equips[:i], es.equipData.Equips[i+1:]...)
				break
			}
		}
		// 放回背包
		err = bagSys.AddItem(ctx, oldEquip.ItemId, 1, 1) // 装备默认绑定
		if err != nil {
			log.Errorf("add old equip to bag failed: %v", err)
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
	es.equipData.Equips = append(es.equipData.Equips, newEquip)

	// 发布事件
	playerRole.Publish(gevent.OnEquipChange, map[string]interface{}{
		"slot":    slot,
		"item_id": itemID,
	})

	log.Infof("EquipItem: PlayerID=%d, Slot=%d, ItemID=%d", playerID, slot, itemID)
	return nil
}

// UnequipItem 卸载装备
func (es *EquipSys) UnequipItem(ctx context.Context, slot uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	if es.equipData == nil || es.equipData.Equips == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	equip := es.findEquipBySlot(slot)
	if equip == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	// 检查背包是否有空间
	bagSys := GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag system not found")
	}

	// 添加到背包
	err = bagSys.AddItem(ctx, equip.ItemId, 1, 1) // 装备默认绑定
	if err != nil {
		return customerr.Wrap(err)
	}

	// 从列表中移除
	for i, e := range es.equipData.Equips {
		if e != nil && e.Slot == slot {
			es.equipData.Equips = append(es.equipData.Equips[:i], es.equipData.Equips[i+1:]...)
			break
		}
	}

	// 发布事件
	playerRole.Publish(gevent.OnEquipChange, map[string]interface{}{
		"slot": slot,
	})

	log.Infof("UnequipItem: Slot=%d, ItemID=%d", slot, equip.ItemId)
	return nil
}

// UpgradeEquip 强化装备
func (es *EquipSys) UpgradeEquip(ctx context.Context, slot uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	playerID := uint(playerRole.GetPlayerRoleId())

	if es.equipData == nil || es.equipData.Equips == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	equip := es.findEquipBySlot(slot)
	if equip == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	// 获取装备强化配置（从单独的配置表）
	upgradeConfig, ok := jsonconf.GetConfigManager().GetEquipUpgradeConfig(equip.ItemId)
	if !ok || upgradeConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "equip upgrade config not found")
	}

	// 检查强化配置
	if len(upgradeConfig.UpgradeCost) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "equip cannot upgrade")
	}

	// 查找当前等级的强化配置
	var upgradeCost *jsonconf.UpgradeCost
	for _, cost := range upgradeConfig.UpgradeCost {
		if cost.Level == equip.Level+1 {
			upgradeCost = cost
			break
		}
	}

	if upgradeCost == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "max upgrade level")
	}

	// 构建消耗列表
	consumeItems := make([]*jsonconf.ItemAmount, 0)

	// 添加金币消耗（使用MoneyTypeGoldCoin）
	if upgradeCost.Gold > 0 {
		consumeItems = append(consumeItems, &jsonconf.ItemAmount{
			ItemType: uint32(protocol.ItemType_ItemTypeMoney),
			ItemId:   uint32(protocol.MoneyType_MoneyTypeGoldCoin), // 使用MoneyTypeGoldCoin
			Count:    int64(upgradeCost.Gold),
		})
	}

	// 添加材料消耗
	if upgradeCost.ItemId > 0 && upgradeCost.Count > 0 {
		consumeItems = append(consumeItems, &jsonconf.ItemAmount{
			ItemType: uint32(protocol.ItemType_ItemTypeMaterial), // 材料类型
			ItemId:   upgradeCost.ItemId,
			Count:    int64(upgradeCost.Count),
		})
	}

	// 检查消耗是否足够
	if len(consumeItems) > 0 {
		if err := playerRole.CheckConsume(ctx, consumeItems); err != nil {
			return customerr.Wrap(err)
		}

		// 扣除消耗
		if err := playerRole.ApplyConsume(ctx, consumeItems); err != nil {
			return customerr.Wrap(err)
		}
	}

	// 更新装备等级
	equip.Level = upgradeCost.Level
	equip.Exp = 0

	// 发布事件
	playerRole.Publish(gevent.OnEquipUpgrade, map[string]interface{}{
		"slot":  slot,
		"level": equip.Level,
	})

	log.Infof("UpgradeEquip: PlayerID=%d, Slot=%d, Level=%d", playerID, slot, equip.Level)
	return nil
}

// GetEquip 获取指定槽位的装备
func (es *EquipSys) GetEquip(slot uint32) *protocol.EquipSt {
	return es.findEquipBySlot(slot)
}

// GetAllEquips 获取所有装备
func (es *EquipSys) GetAllEquips() map[uint32]*protocol.EquipSt {
	result := make(map[uint32]*protocol.EquipSt)
	if es.equipData == nil || es.equipData.Equips == nil {
		return result
	}
	for _, equip := range es.equipData.Equips {
		if equip != nil {
			result[equip.Slot] = equip
		}
	}
	return result
}

// GetEquipData 获取装备数据（用于协议）
func (es *EquipSys) GetEquipData() *protocol.SiEquipData {
	return es.equipData
}

// CalculateEquipAttrs 计算装备属性加成（根据强化等级、品质、星级、阶级）
func (es *EquipSys) CalculateEquipAttrs() map[uint32]int64 {
	attrs := make(map[uint32]int64)

	if es.equipData == nil || es.equipData.Equips == nil {
		return attrs
	}

	for _, equip := range es.equipData.Equips {
		if equip == nil {
			continue
		}
		itemConfig, ok := jsonconf.GetConfigManager().GetItemConfig(equip.ItemId)
		if !ok {
			continue
		}

		// 根据强化等级计算属性值（每级增加10%，使用万分比）
		// 基础倍率10000，每级增加1000
		upgradeMultiplier := PerTenThousandBase + int64(equip.Level-1)*UpgradePerLevelMultiplier

		// 精炼属性（根据精炼等级计算，每级增加5%，使用万分比）
		// 基础倍率10000，每级增加500
		refineMultiplier := PerTenThousandBase
		if equip.RefineLevel > 0 {
			refineMultiplier = PerTenThousandBase + int64(equip.RefineLevel)*RefinePerLevelMultiplier
		}

		// 计算倍率乘积，检查是否会溢出
		baseMultiplier := upgradeMultiplier * refineMultiplier
		if baseMultiplier < 0 || baseMultiplier > math.MaxInt64/PerTenThousandBase {
			log.Errorf("equip attr multiplier overflow: upgrade=%d, refine=%d, base=%d", upgradeMultiplier, refineMultiplier, baseMultiplier)
			continue // 跳过该装备，避免溢出
		}

		// 普通属性（所有装备都有）
		for _, attr := range itemConfig.NormalAttrs {
			if attr == nil {
				continue
			}
			attrValue := int64(attr.Value)
			// 检查计算是否会溢出
			if attrValue > 0 && baseMultiplier > math.MaxInt64/attrValue {
				log.Errorf("equip attr value overflow: value=%d, multiplier=%d", attr.Value, baseMultiplier)
				continue
			}
			// 先乘以倍率，再除以10000（万分比）
			finalValue := (attrValue * baseMultiplier) / PerTenThousandBase / PerTenThousandBase
			attrs[attr.Type] += finalValue
		}

		// 极品属性（根据装备品质判断）
		if itemConfig.Quality >= 2 {
			for _, attr := range itemConfig.RareAttrs {
				if attr == nil {
					continue
				}
				attrValue := int64(attr.Value)
				if attrValue > 0 && baseMultiplier > math.MaxInt64/attrValue {
					log.Errorf("equip rare attr value overflow: value=%d, multiplier=%d", attr.Value, baseMultiplier)
					continue
				}
				finalValue := (attrValue * baseMultiplier) / PerTenThousandBase / PerTenThousandBase
				attrs[attr.Type] += finalValue
			}
		}

		// 星级属性（根据装备星级判断）
		if itemConfig.Star > 0 {
			starMultiplier := PerTenThousandBase + int64(itemConfig.Star)*StarPerLevelMultiplier
			// 检查倍率乘积是否会溢出
			starBaseMultiplier := baseMultiplier * starMultiplier
			if starBaseMultiplier < 0 || starBaseMultiplier > math.MaxInt64/PerTenThousandBase {
				log.Errorf("equip star attr multiplier overflow: base=%d, star=%d, total=%d", baseMultiplier, starMultiplier, starBaseMultiplier)
				continue
			}
			for _, attr := range itemConfig.StarAttrs {
				if attr == nil {
					continue
				}
				attrValue := int64(attr.Value)
				if attrValue > 0 && starBaseMultiplier > math.MaxInt64/attrValue {
					log.Errorf("equip star attr value overflow: value=%d, multiplier=%d", attr.Value, starBaseMultiplier)
					continue
				}
				// 先乘以倍率，再除以10000（万分比）
				finalValue := (attrValue * starBaseMultiplier) / PerTenThousandBase / PerTenThousandBase / PerTenThousandBase
				attrs[attr.Type] += finalValue
			}
		}

		// 品质属性（根据装备品质判断）
		if itemConfig.Quality > 0 {
			qualityMultiplier := PerTenThousandBase + int64(itemConfig.Quality)*QualityPerLevelMultiplier
			// 检查倍率乘积是否会溢出
			qualityBaseMultiplier := baseMultiplier * qualityMultiplier
			if qualityBaseMultiplier < 0 || qualityBaseMultiplier > math.MaxInt64/PerTenThousandBase {
				log.Errorf("equip quality attr multiplier overflow: base=%d, quality=%d, total=%d", baseMultiplier, qualityMultiplier, qualityBaseMultiplier)
				continue
			}
			for _, attr := range itemConfig.QualityAttrs {
				if attr == nil {
					continue
				}
				attrValue := int64(attr.Value)
				if attrValue > 0 && qualityBaseMultiplier > math.MaxInt64/attrValue {
					log.Errorf("equip quality attr value overflow: value=%d, multiplier=%d", attr.Value, qualityBaseMultiplier)
					continue
				}
				// 先乘以倍率，再除以10000（万分比）
				finalValue := (attrValue * qualityBaseMultiplier) / PerTenThousandBase / PerTenThousandBase / PerTenThousandBase
				attrs[attr.Type] += finalValue
			}
		}

		// 阶级属性（根据装备阶级判断）
		if itemConfig.Tier > 0 {
			tierMultiplier := PerTenThousandBase + int64(itemConfig.Tier)*TierPerLevelMultiplier
			// 检查倍率乘积是否会溢出
			tierBaseMultiplier := baseMultiplier * tierMultiplier
			if tierBaseMultiplier < 0 || tierBaseMultiplier > math.MaxInt64/PerTenThousandBase {
				log.Errorf("equip tier attr multiplier overflow: base=%d, tier=%d, total=%d", baseMultiplier, tierMultiplier, tierBaseMultiplier)
				continue
			}
			for _, attr := range itemConfig.TierAttrs {
				if attr == nil {
					continue
				}
				attrValue := int64(attr.Value)
				if attrValue > 0 && tierBaseMultiplier > math.MaxInt64/attrValue {
					log.Errorf("equip tier attr value overflow: value=%d, multiplier=%d", attr.Value, tierBaseMultiplier)
					continue
				}
				// 先乘以倍率，再除以10000（万分比）
				finalValue := (attrValue * tierBaseMultiplier) / PerTenThousandBase / PerTenThousandBase / PerTenThousandBase
				attrs[attr.Type] += finalValue
			}
		}

		// 附魔属性（直接累加）
		if equip.EnchantAttrs != nil && len(equip.EnchantAttrs) > 0 {
			for _, enchantAttr := range equip.EnchantAttrs {
				if enchantAttr != nil {
					attrs[enchantAttr.Type] += enchantAttr.Value
				}
			}
		}
	}

	// 计算套装效果
	setAttrs := es.calculateSetEffects()
	for attrType, attrValue := range setAttrs {
		attrs[attrType] += attrValue
	}

	return attrs
}

// calculateSetEffects 计算套装效果
func (es *EquipSys) calculateSetEffects() map[uint32]int64 {
	setAttrs := make(map[uint32]int64)

	if es.equipData == nil || es.equipData.Equips == nil {
		return setAttrs
	}

	// 统计每个套装的装备数量
	setCounts := make(map[uint32]int)
	for _, equip := range es.equipData.Equips {
		if equip != nil && equip.SetId > 0 {
			setCounts[equip.SetId]++
		}
	}

	// 计算套装效果
	configMgr := jsonconf.GetConfigManager()
	for setId, count := range setCounts {
		setConfig, ok := configMgr.GetEquipSetConfig(setId)
		if !ok || setConfig == nil {
			continue
		}

		// 查找满足条件的套装效果（按件数从大到小排序，取最大的满足条件的）
		var bestEffect *jsonconf.SetEffect
		for _, effect := range setConfig.Effects {
			if effect != nil && uint32(count) >= effect.Count {
				if bestEffect == nil || effect.Count > bestEffect.Count {
					bestEffect = effect
				}
			}
		}

		// 应用套装效果
		if bestEffect != nil && bestEffect.Attrs != nil {
			for _, attr := range bestEffect.Attrs {
				if attr != nil {
					setAttrs[attr.Type] += int64(attr.Value)
				}
			}
		}
	}

	return setAttrs
}

// CalculateAttrs 计算装备系统的属性（实现IAttrCalculator接口）
func (es *EquipSys) CalculateAttrs(ctx context.Context) []*protocol.AttrSt {
	// 使用已有的CalculateEquipAttrs方法
	attrs := es.CalculateEquipAttrs()
	if len(attrs) == 0 {
		return nil
	}

	// 转换为protocol.AttrSt格式
	result := make([]*protocol.AttrSt, 0, len(attrs))
	for attrType, attrValue := range attrs {
		result = append(result, &protocol.AttrSt{
			Type:  attrType,
			Value: attrValue,
		})
	}

	return result
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysEquip), func() iface.ISystem {
		return NewEquipSys()
	})
	gevent.SubscribePlayerEvent(gevent.OnEquipChange, func(ctx context.Context, ev *event.Event) {
		// 装备变更时标记属性系统需要重算
		attrSys := GetAttrSys(ctx)
		if attrSys != nil {
			attrSys.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
		}
	})
	gevent.SubscribePlayerEvent(gevent.OnEquipUpgrade, func(ctx context.Context, ev *event.Event) {
		// 装备升级时标记属性系统需要重算
		attrSys := GetAttrSys(ctx)
		if attrSys != nil {
			attrSys.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
		}
	})
}

// RefineEquip 精炼装备（提升品质）
func (es *EquipSys) RefineEquip(ctx context.Context, slot uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	equip := es.findEquipBySlot(slot)
	if equip == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	// 获取精炼配置
	refineConfig, ok := jsonconf.GetConfigManager().GetEquipRefineConfig(equip.ItemId)
	if !ok || refineConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "equip refine config not found")
	}

	// 查找下一级精炼配置
	targetRefineLevel := equip.RefineLevel + 1
	var refineCost *jsonconf.RefineCost
	for _, cost := range refineConfig.RefineCosts {
		if cost != nil && cost.RefineLevel == targetRefineLevel {
			refineCost = cost
			break
		}
	}

	if refineCost == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "max refine level")
	}

	// 检查消耗
	if len(refineCost.Consume) > 0 {
		if err := playerRole.CheckConsume(ctx, refineCost.Consume); err != nil {
			return customerr.Wrap(err)
		}

		// 扣除消耗
		if err := playerRole.ApplyConsume(ctx, refineCost.Consume); err != nil {
			return customerr.Wrap(err)
		}
	}

	// 提升精炼等级
	equip.RefineLevel = targetRefineLevel

	// 如果配置了品质提升，更新装备品质（需要从ItemConfig更新，但ItemConfig是只读的，这里只记录精炼等级）
	// 品质提升在属性计算时通过精炼等级体现

	// 标记属性系统需要重算
	attrSys := GetAttrSys(ctx)
	if attrSys != nil {
		attrSys.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
	}

	log.Infof("RefineEquip: Slot=%d, RefineLevel=%d", slot, equip.RefineLevel)
	return nil
}

// EnchantEquip 附魔装备（添加特殊属性）
func (es *EquipSys) EnchantEquip(ctx context.Context, slot uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	equip := es.findEquipBySlot(slot)
	if equip == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no equip in slot")
	}

	// 获取附魔配置
	enchantConfig, ok := jsonconf.GetConfigManager().GetEquipEnchantConfig(equip.ItemId)
	if !ok || enchantConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "equip enchant config not found")
	}

	// 检查附魔数量上限
	if equip.EnchantAttrs != nil && uint32(len(equip.EnchantAttrs)) >= enchantConfig.MaxEnchantCount {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "max enchant count reached")
	}

	// 检查消耗
	if len(enchantConfig.EnchantCost) > 0 {
		if err := playerRole.CheckConsume(ctx, enchantConfig.EnchantCost); err != nil {
			return customerr.Wrap(err)
		}

		// 扣除消耗
		if err := playerRole.ApplyConsume(ctx, enchantConfig.EnchantCost); err != nil {
			return customerr.Wrap(err)
		}
	}

	// 随机选择一个附魔属性（如果有多个）
	if len(enchantConfig.EnchantAttrs) > 0 {
		if equip.EnchantAttrs == nil {
			equip.EnchantAttrs = make([]*protocol.AttrSt, 0)
		}

		// 简单实现：选择第一个附魔属性（后续可以改为随机选择）
		selectedAttr := enchantConfig.EnchantAttrs[0]
		if selectedAttr != nil {
			equip.EnchantAttrs = append(equip.EnchantAttrs, &protocol.AttrSt{
				Type:  selectedAttr.Type,
				Value: int64(selectedAttr.Value),
			})
		}
	}

	// 标记属性系统需要重算
	attrSys := GetAttrSys(ctx)
	if attrSys != nil {
		attrSys.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
	}

	log.Infof("EnchantEquip: Slot=%d, EnchantCount=%d", slot, len(equip.EnchantAttrs))
	return nil
}
