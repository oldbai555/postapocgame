package system

import (
	"context"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	gameattrcalc "postapocgame/server/service/gameserver/internel/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/di"
)

// CalculateAttrs 计算等级系统的属性（实现属性计算器接口）
func (a *LevelSystemAdapter) CalculateAttrs(ctx context.Context) []*protocol.AttrSt {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil
	}
	if binaryData.LevelData == nil {
		return nil
	}

	// 从配置表获取等级属性
	levelAttrs := jsonconf.GetConfigManager().GetLevelAttrs(binaryData.LevelData.Level)
	if len(levelAttrs) == 0 {
		return nil
	}

	// 转换为protocol.AttrSt格式
	result := make([]*protocol.AttrSt, 0, len(levelAttrs))
	for attrType, attrValue := range levelAttrs {
		result = append(result, &protocol.AttrSt{
			Type:  attrType,
			Value: int64(attrValue),
		})
	}

	return result
}

// levelAddRateCalculator 等级加成计算器
type levelAddRateCalculator struct{}

// CalculateAddRate 计算等级加成属性
func (c *levelAddRateCalculator) CalculateAddRate(ctx context.Context, _ *icalc.FightAttrCalc) []*protocol.AttrSt {
	levelSys := GetLevelSys(ctx)
	if levelSys == nil {
		return nil
	}
	level, err := levelSys.GetLevel(ctx)
	if err != nil || level == 0 {
		return nil
	}
	cfg := jsonconf.GetConfigManager().GetAttrAddRateConfig()
	results := make([]*protocol.AttrSt, 0, 2)
	if cfg.Level.HPRegenPerLevel > 0 {
		regenHP := int64(level) * cfg.Level.HPRegenPerLevel
		if regenHP > 0 {
			results = append(results, &protocol.AttrSt{
				Type:  uint32(attrdef.AttrHPRegen),
				Value: regenHP,
			})
		}
	}
	if cfg.Level.MPRegenPerLevel > 0 {
		regenMP := int64(level) * cfg.Level.MPRegenPerLevel
		if regenMP > 0 {
			results = append(results, &protocol.AttrSt{
				Type:  uint32(attrdef.AttrMPRegen),
				Value: regenMP,
			})
		}
	}
	if len(results) == 0 {
		return nil
	}
	return results
}

// EnsureISysAttrCalculator 确保 LevelSystemAdapter 实现属性计算器接口
var _ gameattrcalc.Calculator = (*LevelSystemAdapter)(nil)

// EnsureISysAddRateCalculator 确保 levelAddRateCalculator 实现加成计算器接口
var _ gameattrcalc.AddRateCalculator = (*levelAddRateCalculator)(nil)
