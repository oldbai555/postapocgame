// skill_damage_calculator.go 负责所有基础伤害、治疗与命中计算逻辑。
package skill

import (
	"math/rand"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	iface2 "postapocgame/server/service/gameserver/internel/dungeonactor/iface"
)

// DamageCalculator 伤害计算器
type DamageCalculator struct {
}

// NewDamageCalculator 创建伤害计算器
func NewDamageCalculator() *DamageCalculator {
	return &DamageCalculator{}
}

// Calculate 计算伤害（完整伤害公式）
// 返回: 伤害值, 是否暴击, 是否闪避
func (dc *DamageCalculator) Calculate(attacker, defender iface2.IEntity, skillId uint32) (int64, bool, bool) {
	// 获取攻击者和防御者的属性系统
	attackerAttrSys := attacker.GetAttrSys()
	defenderAttrSys := defender.GetAttrSys()

	// 检查闪避
	if dc.CheckDodge(attackerAttrSys, defenderAttrSys) {
		return 0, false, true
	}

	// 获取基础属性
	attackerAttack := attackerAttrSys.GetAttrValue(attrdef.Attack)
	defenderDefense := defenderAttrSys.GetAttrValue(attrdef.Defense)

	// 基础伤害公式: (攻击力 * 技能倍率) - (防御力 * 防御减免系数)
	baseDamage := float64(attackerAttack)

	// 技能伤害加成
	skillMultiplier := 1.0
	if skillId > 0 {
		skillConfig := jsonconf.GetConfigManager().GetSkillConfig(skillId)
		if skillConfig != nil {
			// 技能基础伤害 + 攻击力百分比
			skillBaseDamage := float64(skillConfig.Damage)
			skillMultiplier = 1.0 + float64(skillConfig.Damage)/100.0 // 假设Damage是百分比
			baseDamage = float64(attackerAttack)*skillMultiplier + skillBaseDamage
		} else {
			// 默认技能倍率
			baseDamage *= 1.5
		}
	}

	// 防御减免: 防御力 / (防御力 + 攻击者等级 * 10)
	attackerLevel := attacker.GetLevel()
	defenseReduce := float64(defenderDefense) / (float64(defenderDefense) + float64(attackerLevel)*10.0)
	if defenseReduce > 0.8 {
		defenseReduce = 0.8 // 最大减免80%
	}

	// 计算最终伤害
	finalDamage := baseDamage * (1.0 - defenseReduce)

	// 检查暴击
	isCrit := dc.checkCrit(attackerAttrSys)
	if isCrit {
		critDamage := attackerAttrSys.GetAttrValue(attrdef.CritDamage)
		critMultiplier := 2.0 + float64(critDamage)/10000.0 // 基础2倍 + 暴击伤害加成
		finalDamage *= critMultiplier
	}

	// 随机浮动 ±10%
	randomFactor := 0.9 + rand.Float64()*0.2
	finalDamage *= randomFactor

	// 最小伤害保证
	if finalDamage < 1 {
		finalDamage = 1
	}

	return int64(finalDamage), isCrit, false
}

// checkCrit 检查是否暴击
func (dc *DamageCalculator) checkCrit(attrSys iface2.IAttrSys) bool {
	critRate := attrSys.GetAttrValue(attrdef.CritRate)
	// critRate是万分比，转换为概率
	prob := float64(critRate) / 10000.0
	return rand.Float64() < prob
}

// CheckDodge 检查是否闪避
func (dc *DamageCalculator) CheckDodge(attackerAttrSys, defenderAttrSys iface2.IAttrSys) bool {
	// 获取闪避率和命中率
	defenderDodgeRate := defenderAttrSys.GetAttrValue(attrdef.DodgeRate)
	attackerHitRate := attackerAttrSys.GetAttrValue(attrdef.HitRate)

	// 实际闪避率 = 防御者闪避率 - 攻击者命中率
	actualDodgeRate := float64(defenderDodgeRate) - float64(attackerHitRate)
	if actualDodgeRate < 0 {
		actualDodgeRate = 0
	}

	// 转换为概率（万分比）
	prob := actualDodgeRate / 10000.0

	// 速度差影响闪避
	attackerSpeed := attackerAttrSys.GetAttrValue(attrdef.Speed)
	defenderSpeed := defenderAttrSys.GetAttrValue(attrdef.Speed)
	if defenderSpeed > attackerSpeed {
		speedDiff := float64(defenderSpeed - attackerSpeed)
		prob += speedDiff * 0.0001 // 每点速度差增加0.01%闪避
	}

	// 最大闪避率50%
	if prob > 0.5 {
		prob = 0.5
	}

	return rand.Float64() < prob
}

// CalculateSkillDamage 计算技能伤害
func (dc *DamageCalculator) CalculateSkillDamage(attacker, defender iface2.IEntity, skillId uint32) int64 {
	damage, _, _ := dc.Calculate(attacker, defender, skillId)
	return damage
}

// CalculateHeal 计算治疗量
func (dc *DamageCalculator) CalculateHeal(caster iface2.IEntity, target iface2.IEntity, skillId uint32) int64 {
	casterAttrSys := caster.GetAttrSys()

	// 基础治疗 = 攻击力 * 基础倍率
	attack := casterAttrSys.GetAttrValue(attrdef.Attack)
	baseHeal := float64(attack) * 0.5

	// 技能加成
	if skillId > 0 {
		skillConfig := jsonconf.GetConfigManager().GetSkillConfig(skillId)
		if skillConfig != nil {
			// 从技能效果中查找治疗效果
			for _, effect := range skillConfig.Effects {
				if effect.Type == 2 { // 治疗效果
					baseHeal = float64(effect.Value)
					if effect.Rate > 0 {
						baseHeal *= float64(effect.Rate)
					}
					break
				}
			}
		}
	}

	// 随机浮动 ±10%
	randomFactor := 0.9 + rand.Float64()*0.2
	finalHeal := baseHeal * randomFactor

	return int64(finalHeal)
}
