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

// FightAttrCalc 高级战斗属性计算器（支持拷贝、合并、相减等操作）
type FightAttrCalc struct {
	values [CombatAttrCount]attrdef.AttrValue
}

// NewFightAttrCalc 创建 FightAttrCalc
func NewFightAttrCalc() *FightAttrCalc {
	return &FightAttrCalc{}
}

// Reset 清空所有属性
func (calc *FightAttrCalc) Reset() {
	copy(calc.values[:], ZeroCombatAttrs[:])
}

// Copy 拷贝另一份属性
func (calc *FightAttrCalc) Copy(src *FightAttrCalc) {
	if src == nil {
		calc.Reset()
		return
	}
	copy(calc.values[:], src.values[:])
}

// Clone 复制出新的计算器
func (calc *FightAttrCalc) Clone() *FightAttrCalc {
	clone := NewFightAttrCalc()
	clone.Copy(calc)
	return clone
}

// SetValue 设置属性值
func (calc *FightAttrCalc) SetValue(attrType attrdef.AttrType, value attrdef.AttrValue) bool {
	if !attrdef.IsCombatAttr(attrType) {
		return false
	}
	idx := attrType - attrdef.CombatAttrBegin
	changed := calc.values[idx] != value
	calc.values[idx] = value
	return changed
}

// GetValue 获取属性值
func (calc *FightAttrCalc) GetValue(attrType attrdef.AttrType) attrdef.AttrValue {
	if !attrdef.IsCombatAttr(attrType) {
		return 0
	}
	return calc.values[attrType-attrdef.CombatAttrBegin]
}

// AddValue 属性值累加
func (calc *FightAttrCalc) AddValue(attrType attrdef.AttrType, delta attrdef.AttrValue) {
	if !attrdef.IsCombatAttr(attrType) {
		return
	}
	idx := attrType - attrdef.CombatAttrBegin
	calc.values[idx] += delta
}

// AddCalc 按属性累加另一份属性
func (calc *FightAttrCalc) AddCalc(other *FightAttrCalc) {
	if other == nil {
		return
	}
	for i, v := range other.values {
		calc.values[i] += v
	}
}

// SubCalc 按属性扣减另一份属性
func (calc *FightAttrCalc) SubCalc(other *FightAttrCalc) {
	if other == nil {
		return
	}
	for i, v := range other.values {
		calc.values[i] -= v
	}
}

// DoRange 遍历所有属性
func (calc *FightAttrCalc) DoRange(cb func(attrType attrdef.AttrType, value attrdef.AttrValue)) {
	for idx, v := range calc.values {
		if v == 0 {
			continue
		}
		cb(attrdef.AttrType(idx)+attrdef.CombatAttrBegin, v)
	}
}
