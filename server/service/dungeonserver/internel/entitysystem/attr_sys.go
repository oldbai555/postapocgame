/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 属性系统实现
**/

package entitysystem

import (
	"postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

var _ iface.IAttrSys = (*AttrSys)(nil)

// AttrSys 属性系统
type AttrSys struct {
	// 战斗属性计算器
	combatCalc *attrcalc.CombatAttrCalc

	// 非战斗属性计算器
	extraCalc *attrcalc.ExtraAttrCalc
}

// NewAttrSys 创建属性系统
func NewAttrSys() *AttrSys {
	return &AttrSys{
		combatCalc: attrcalc.NewCombatAttrCalc(),
		extraCalc:  attrcalc.NewExtraAttrCalc(),
	}
}

// GetAttrValue 获取属性值
func (as *AttrSys) GetAttrValue(attrType attrdef.AttrType) attrdef.AttrValue {
	if attrdef.IsCombatAttr(attrType) {
		return as.combatCalc.GetValue(attrType)
	}

	if attrdef.IsExtraAttr(attrType) {
		return as.extraCalc.GetValue(attrType)
	}

	return 0
}

func (as *AttrSys) SetAttrValue(attrType attrdef.AttrType, value attrdef.AttrValue) {
	if attrdef.IsCombatAttr(attrType) {
		as.combatCalc.SetValue(attrType, value)
		return
	}

	if attrdef.IsExtraAttr(attrType) {
		as.extraCalc.SetValue(attrType, value)
		return
	}
}

func (as *AttrSys) AddAttrValue(attrType attrdef.AttrType, delta attrdef.AttrValue) {
	if attrdef.IsCombatAttr(attrType) {
		as.combatCalc.AddValue(attrType, delta)
		return
	}

	if attrdef.IsExtraAttr(attrType) {
		as.extraCalc.AddValue(attrType, delta)
		return
	}
}

// GetAllCombatAttrs 获取所有战斗属性
func (as *AttrSys) GetAllCombatAttrs() map[attrdef.AttrType]attrdef.AttrValue {
	return as.combatCalc.GetAll()
}

// GetAllExtraAttrs 获取所有非战斗属性
func (as *AttrSys) GetAllExtraAttrs() map[attrdef.AttrType]attrdef.AttrValue {
	return as.extraCalc.GetAll()
}

// ResetCombatAttrs 重置战斗属性
func (as *AttrSys) ResetCombatAttrs() {
	as.combatCalc.Reset()
}

// ResetExtraAttrs 重置非战斗属性
func (as *AttrSys) ResetExtraAttrs() {
	as.extraCalc.Reset()
}

// ResetAll 重置所有属性
func (as *AttrSys) ResetAll() {
	as.combatCalc.Reset()
	as.extraCalc.Reset()
}

// BatchSetAttrs 批量设置属性
func (as *AttrSys) BatchSetAttrs(attrs map[attrdef.AttrType]attrdef.AttrValue) {
	for attrType, value := range attrs {
		if attrdef.IsCombatAttr(attrType) {
			as.combatCalc.SetValue(attrType, value)
		} else if attrdef.IsExtraAttr(attrType) {
			as.extraCalc.SetValue(attrType, value)
		}
	}

	return
}

// BatchAddAttrs 批量增加属性
func (as *AttrSys) BatchAddAttrs(attrs map[attrdef.AttrType]attrdef.AttrValue) {
	for attrType, delta := range attrs {
		if attrdef.IsCombatAttr(attrType) {
			as.combatCalc.AddValue(attrType, delta)
		} else if attrdef.IsExtraAttr(attrType) {
			as.extraCalc.AddValue(attrType, delta)
		}
	}

	return
}

// GetHP 获取当前生命值
func (as *AttrSys) GetHP() int64 {
	val := as.GetAttrValue(attrdef.AttrHP)
	return int64(val)
}

// GetMaxHP 获取最大生命值
func (as *AttrSys) GetMaxHP() int64 {
	val := as.GetAttrValue(attrdef.AttrMaxHP)
	return int64(val)
}

// GetMP 获取当前魔法值
func (as *AttrSys) GetMP() int64 {
	val := as.GetAttrValue(attrdef.AttrMP)
	return int64(val)
}

// GetMaxMP 获取最大魔法值
func (as *AttrSys) GetMaxMP() int64 {
	val := as.GetAttrValue(attrdef.AttrMaxMP)
	return int64(val)
}

// GetAttack 获取攻击力
func (as *AttrSys) GetAttack() int64 {
	val := as.GetAttrValue(attrdef.AttrAttack)
	return int64(val)
}

// GetDefense 获取防御力
func (as *AttrSys) GetDefense() int64 {
	val := as.GetAttrValue(attrdef.AttrDefense)
	return int64(val)
}

// GetSpeed 获取速度
func (as *AttrSys) GetSpeed() int64 {
	val := as.GetAttrValue(attrdef.AttrSpeed)
	return int64(val)
}

// GetCritRate 获取暴击率（万分比）
func (as *AttrSys) GetCritRate() int64 {
	val := as.GetAttrValue(attrdef.AttrCritRate)
	return int64(val)
}

// GetDodgeRate 获取闪避率（万分比）
func (as *AttrSys) GetDodgeRate() int64 {
	val := as.GetAttrValue(attrdef.AttrDodgeRate)
	return int64(val)
}

// SetHP 设置当前生命值
func (as *AttrSys) SetHP(hp int64) {
	// 限制HP不超过MaxHP
	maxHP := as.GetMaxHP()
	if hp > maxHP {
		hp = maxHP
	}
	if hp < 0 {
		hp = 0
	}
	as.SetAttrValue(attrdef.AttrHP, hp)
}

// SetMP 设置当前魔法值
func (as *AttrSys) SetMP(mp int64) {
	// 限制MP不超过MaxMP
	maxMP := as.GetMaxMP()
	if mp > maxMP {
		mp = maxMP
	}
	if mp < 0 {
		mp = 0
	}
	as.SetAttrValue(attrdef.AttrMP, attrdef.AttrValue(mp))
}

// AddHP 增加生命值
func (as *AttrSys) AddHP(delta int64) {
	currentHP := as.GetHP()
	as.SetHP(currentHP + delta)
}

// AddMP 增加魔法值
func (as *AttrSys) AddMP(delta int64) {
	currentMP := as.GetMP()
	as.SetMP(currentMP + delta)
}
