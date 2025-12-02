package system

import (
	"context"
	"math"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	gameattrcalc "postapocgame/server/service/gameserver/internel/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

const (
	// PerTenThousandBase 万分比基数（10000 = 100%）
	PerTenThousandBase int64 = 10000
	// UpgradePerLevelMultiplier 强化等级每级增加的倍率（万分比：1000 = 10%）
	UpgradePerLevelMultiplier int64 = 1000
	// RefinePerLevelMultiplier 精炼等级每级增加的倍率（万分比：500 = 5%）
	RefinePerLevelMultiplier int64 = 500
	// StarPerLevelMultiplier 星级每星增加的倍率（万分比：1000 = 10%）
	StarPerLevelMultiplier int64 = 1000
	// QualityPerLevelMultiplier 品质每级增加的倍率（万分比：1000 = 10%）
	QualityPerLevelMultiplier int64 = 1000
	// TierPerLevelMultiplier 阶级每级增加的倍率（万分比：1000 = 10%）
	TierPerLevelMultiplier int64 = 1000
)

// CalculateAttrs 计算装备系统的属性（实现属性计算器接口）
func (a *EquipSystemAdapter) CalculateAttrs(ctx context.Context) []*protocol.AttrSt {
	equipData := a.GetEquipData(ctx)
	if equipData == nil || equipData.Equips == nil {
		return nil
	}

	// 计算所有装备的属性
	attrs := make(map[uint32]int64)
	configMgr := di.GetContainer().ConfigGateway()

	for _, equip := range equipData.Equips {
		if equip == nil {
			continue
		}
		// 计算单个装备的属性（包括基础属性、强化、精炼、附魔等）
		equipAttrs := a.calculateEquipAttrs(ctx, equip, configMgr)
		for attrType, attrValue := range equipAttrs {
			attrs[attrType] += attrValue
		}
	}

	// 计算套装效果
	setAttrs := a.calculateSetEffects(equipData, configMgr)
	for attrType, attrValue := range setAttrs {
		attrs[attrType] += attrValue
	}

	// 转换为 protocol.AttrSt 格式
	if len(attrs) == 0 {
		return nil
	}
	result := make([]*protocol.AttrSt, 0, len(attrs))
	for attrType, attrValue := range attrs {
		result = append(result, &protocol.AttrSt{
			Type:  attrType,
			Value: attrValue,
		})
	}
	return result
}

// calculateEquipAttrs 计算单个装备的属性
func (a *EquipSystemAdapter) calculateEquipAttrs(ctx context.Context, equip *protocol.EquipSt, configMgr interfaces.ConfigManager) map[uint32]int64 {
	attrs := make(map[uint32]int64)

	// 获取装备配置
	itemConfigRaw, ok := configMgr.GetItemConfig(equip.ItemId)
	if !ok {
		return attrs
	}
	itemConfig, ok := itemConfigRaw.(*jsonconf.ItemConfig)
	if !ok {
		return attrs
	}

	// 计算基础倍率（强化、精炼）
	baseMultiplier := PerTenThousandBase
	if equip.Level > 0 {
		baseMultiplier += int64(equip.Level) * UpgradePerLevelMultiplier
	}
	if equip.RefineLevel > 0 {
		baseMultiplier += int64(equip.RefineLevel) * RefinePerLevelMultiplier
	}

	// 普通属性
	for _, attr := range itemConfig.NormalAttrs {
		if attr == nil {
			continue
		}
		attrValue := int64(attr.Value)
		if attrValue > 0 && baseMultiplier > math.MaxInt64/attrValue {
			log.Errorf("equip attr value overflow: value=%d, multiplier=%d", attr.Value, baseMultiplier)
			continue
		}
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

	// 星级属性
	if itemConfig.Star > 0 {
		starMultiplier := PerTenThousandBase + int64(itemConfig.Star)*StarPerLevelMultiplier
		starBaseMultiplier := baseMultiplier * starMultiplier
		if starBaseMultiplier < 0 || starBaseMultiplier > math.MaxInt64/PerTenThousandBase {
			log.Errorf("equip star attr multiplier overflow: base=%d, star=%d", baseMultiplier, starMultiplier)
		} else {
			for _, attr := range itemConfig.StarAttrs {
				if attr == nil {
					continue
				}
				attrValue := int64(attr.Value)
				if attrValue > 0 && starBaseMultiplier > math.MaxInt64/attrValue {
					log.Errorf("equip star attr value overflow: value=%d, multiplier=%d", attr.Value, starBaseMultiplier)
					continue
				}
				finalValue := (attrValue * starBaseMultiplier) / PerTenThousandBase / PerTenThousandBase / PerTenThousandBase
				attrs[attr.Type] += finalValue
			}
		}
	}

	// 品质属性
	if itemConfig.Quality > 0 {
		qualityMultiplier := PerTenThousandBase + int64(itemConfig.Quality)*QualityPerLevelMultiplier
		qualityBaseMultiplier := baseMultiplier * qualityMultiplier
		if qualityBaseMultiplier < 0 || qualityBaseMultiplier > math.MaxInt64/PerTenThousandBase {
			log.Errorf("equip quality attr multiplier overflow: base=%d, quality=%d", baseMultiplier, qualityMultiplier)
		} else {
			for _, attr := range itemConfig.QualityAttrs {
				if attr == nil {
					continue
				}
				attrValue := int64(attr.Value)
				if attrValue > 0 && qualityBaseMultiplier > math.MaxInt64/attrValue {
					log.Errorf("equip quality attr value overflow: value=%d, multiplier=%d", attr.Value, qualityBaseMultiplier)
					continue
				}
				finalValue := (attrValue * qualityBaseMultiplier) / PerTenThousandBase / PerTenThousandBase / PerTenThousandBase
				attrs[attr.Type] += finalValue
			}
		}
	}

	// 阶级属性
	if itemConfig.Tier > 0 {
		tierMultiplier := PerTenThousandBase + int64(itemConfig.Tier)*TierPerLevelMultiplier
		tierBaseMultiplier := baseMultiplier * tierMultiplier
		if tierBaseMultiplier < 0 || tierBaseMultiplier > math.MaxInt64/PerTenThousandBase {
			log.Errorf("equip tier attr multiplier overflow: base=%d, tier=%d", baseMultiplier, tierMultiplier)
		} else {
			for _, attr := range itemConfig.TierAttrs {
				if attr == nil {
					continue
				}
				attrValue := int64(attr.Value)
				if attrValue > 0 && tierBaseMultiplier > math.MaxInt64/attrValue {
					log.Errorf("equip tier attr value overflow: value=%d, multiplier=%d", attr.Value, tierBaseMultiplier)
					continue
				}
				finalValue := (attrValue * tierBaseMultiplier) / PerTenThousandBase / PerTenThousandBase / PerTenThousandBase
				attrs[attr.Type] += finalValue
			}
		}
	}

	// 附魔属性（直接累加）
	if len(equip.EnchantAttrs) > 0 {
		for _, enchantAttr := range equip.EnchantAttrs {
			if enchantAttr != nil {
				attrs[enchantAttr.Type] += enchantAttr.Value
			}
		}
	}

	return attrs
}

// calculateSetEffects 计算套装效果
func (a *EquipSystemAdapter) calculateSetEffects(equipData *protocol.SiEquipData, configMgr interfaces.ConfigManager) map[uint32]int64 {
	setAttrs := make(map[uint32]int64)
	if equipData == nil || equipData.Equips == nil {
		return setAttrs
	}

	// 统计每个套装的装备数量
	setCounts := make(map[uint32]int)
	for _, equip := range equipData.Equips {
		if equip != nil && equip.SetId > 0 {
			setCounts[equip.SetId]++
		}
	}

	// 计算套装效果
	for setId, count := range setCounts {
		setConfigRaw, ok := configMgr.GetEquipSetConfig(setId)
		if !ok {
			continue
		}
		setConfig, ok := setConfigRaw.(*jsonconf.EquipSetConfig)
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

// EnsureISysAttrCalculator 确保 EquipSystemAdapter 实现属性计算器接口
var _ gameattrcalc.Calculator = (*EquipSystemAdapter)(nil)
