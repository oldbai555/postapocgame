package skill

import (
	"math/rand"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

// DamageCalculator 伤害计算器
type DamageCalculator struct {
}

// NewDamageCalculator 创建伤害计算器
func NewDamageCalculator() *DamageCalculator {
	return &DamageCalculator{}
}

// Calculate 计算伤害
// 返回: 伤害值, 是否暴击, 是否闪避
func (dc *DamageCalculator) Calculate(attacker, defender iface.IEntity, skillId uint32) (int64, bool, bool) {
	// 获取攻击者和防御者的属性
	// TODO: 从实体中获取实际属性
	attackerAttr := dc.GetEntityAttr(attacker)
	defenderAttr := dc.GetEntityAttr(defender)

	// 检查闪避
	if dc.CheckDodge(attackerAttr, defenderAttr) {
		return 0, false, true
	}

	// 基础伤害 = 攻击力 - 防御力
	baseDamage := float64(attackerAttr.Attack) - float64(defenderAttr.Defense)*0.5
	if baseDamage < 1 {
		baseDamage = 1
	}

	// 技能伤害加成
	if skillId > 0 {
		// TODO: 从技能配置中获取伤害加成
		baseDamage *= 1.5
	}

	// 检查暴击
	isCrit := dc.checkCrit(attackerAttr)
	if isCrit {
		baseDamage *= 2.0 // 暴击伤害 = 基础伤害 * 2
	}

	// 随机浮动 ±10%
	randomFactor := 0.9 + rand.Float64()*0.2
	finalDamage := baseDamage * randomFactor

	if finalDamage < 1 {
		finalDamage = 1
	}

	return int64(finalDamage), isCrit, false
}

// EntityAttr 实体属性
type EntityAttr struct {
	HP        uint32
	MP        uint32
	Attack    uint32
	Defense   uint32
	Speed     uint32
	CritRate  float32 // 暴击率
	DodgeRate float32 // 闪避率
}

func (dc *DamageCalculator) GetEntityAttr(ve iface.IEntity) *EntityAttr {
	attr := &EntityAttr{
		HP:        1000,
		MP:        500,
		Attack:    100,
		Defense:   50,
		Speed:     10,
		CritRate:  0.1,  // 10% 暴击率
		DodgeRate: 0.05, // 5% 闪避率
	}

	// 根据实体类型获取不同属性
	switch ve.GetEntityType() {
	case custom_id.EntityTypeRole:
		// 从角色信息获取属性
		// TODO: 根据等级和装备计算属性
		attr.Attack = 100 + ve.GetLevel()*10
		attr.Defense = 50 + ve.GetLevel()*5
	case custom_id.EntityTypeMonster:
		// TODO: 从怪物配置获取属性
		attr.HP = 2000
		attr.Attack = 80
		attr.Defense = 40
		attr.CritRate = 0.05
		attr.DodgeRate = 0.02
	}

	return attr
}

// checkCrit 检查是否暴击
func (dc *DamageCalculator) checkCrit(attr *EntityAttr) bool {
	return rand.Float32() < attr.CritRate
}

// CheckDodge 检查是否闪避
func (dc *DamageCalculator) CheckDodge(attackerAttr, defenderAttr *EntityAttr) bool {
	// 闪避率受攻击者命中和防御者闪避影响
	dodgeRate := defenderAttr.DodgeRate

	// 速度差影响闪避
	if defenderAttr.Speed > attackerAttr.Speed {
		speedDiff := float32(defenderAttr.Speed - attackerAttr.Speed)
		dodgeRate += speedDiff * 0.001 // 每点速度差增加0.1%闪避
	}

	// 最大闪避率50%
	if dodgeRate > 0.5 {
		dodgeRate = 0.5
	}

	return rand.Float32() < dodgeRate
}

// CalculateSkillDamage 计算技能伤害
func (dc *DamageCalculator) CalculateSkillDamage(attacker, defender iface.IEntity, skillId uint32) int64 {
	damage, _, _ := dc.Calculate(attacker, defender, skillId)
	return damage
}

// CalculateHeal 计算治疗量
func (dc *DamageCalculator) CalculateHeal(caster iface.IEntity, target iface.IEntity, skillId uint32) int64 {
	// TODO: 根据技能配置和施法者属性计算治疗量
	casterAttr := dc.GetEntityAttr(caster)

	// 基础治疗 = 攻击力 * 0.5
	baseHeal := float32(casterAttr.Attack) * 0.5

	// 技能加成
	// TODO: 从技能配置中获取治疗加成
	baseHeal *= 1.2

	return int64(baseHeal)
}
