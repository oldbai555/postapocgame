/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 属性计算工具包
**/

package attrcalc

import (
	"postapocgame/server/internal/attrdef"
)

const (
	CombatAttrCount = attrdef.CombatAttrEnd - attrdef.CombatAttrBegin + 1
	ExtraAttrCount  = attrdef.ExtraAttrEnd - attrdef.ExtraAttrBegin + 1
)

var (
	ZeroCombatAttrs = [CombatAttrCount]attrdef.AttrValue{}
	ZeroExtraAttrs  = [ExtraAttrCount]attrdef.AttrValue{}
)

// CombatAttrCalc 战斗属性计算器
type CombatAttrCalc struct {
	combatAttr [CombatAttrCount]attrdef.AttrValue
}

// NewCombatAttrCalc 创建战斗属性计算器
func NewCombatAttrCalc() *CombatAttrCalc {
	return &CombatAttrCalc{}
}

// Reset 重置所有属性
func (calc *CombatAttrCalc) Reset() {
	copy(calc.combatAttr[:], ZeroCombatAttrs[:])
}

// SetValue 设置属性值
func (calc *CombatAttrCalc) SetValue(attrType attrdef.AttrType, value attrdef.AttrValue) {
	if !attrdef.IsCombatAttr(attrType) {
		return
	}
	calc.combatAttr[attrType-attrdef.CombatAttrBegin] = value
	return
}

// GetValue 获取属性值
func (calc *CombatAttrCalc) GetValue(attrType attrdef.AttrType) attrdef.AttrValue {
	if !attrdef.IsCombatAttr(attrType) {
		return 0
	}
	return calc.combatAttr[attrType-attrdef.CombatAttrBegin]
}

// AddValue 增加属性值
func (calc *CombatAttrCalc) AddValue(attrType attrdef.AttrType, delta attrdef.AttrValue) {
	if !attrdef.IsCombatAttr(attrType) {
		return
	}
	calc.combatAttr[attrType-attrdef.CombatAttrBegin] += delta
	return
}

// DoRange 遍历所有属性
func (calc *CombatAttrCalc) DoRange(cb func(attrType attrdef.AttrType, value attrdef.AttrValue)) {
	for i, v := range calc.combatAttr {
		if v != 0 {
			cb(attrdef.AttrType(i)+attrdef.CombatAttrBegin, v)
		}
	}
}

// GetAll 获取所有属性
func (calc *CombatAttrCalc) GetAll() map[attrdef.AttrType]attrdef.AttrValue {
	result := make(map[attrdef.AttrType]attrdef.AttrValue)
	calc.DoRange(func(attrType attrdef.AttrType, value attrdef.AttrValue) {
		result[attrType] = value
	})
	return result
}
