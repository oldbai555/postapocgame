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
	CombatAttrCount = attrdef.FightAttrEnd - attrdef.FightAttrBegin + 1
	ExtraAttrCount  = attrdef.ExtraAttrEnd - attrdef.ExtraAttrBegin + 1
)

var (
	ZeroCombatAttrs = [CombatAttrCount]int64{}
	ZeroExtraAttrs  = [ExtraAttrCount]int64{}
)

// FightAttrCalc 高级战斗属性计算器（支持拷贝、合并、相减等操作）
type FightAttrCalc struct {
	values [CombatAttrCount]int64
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
func (calc *FightAttrCalc) SetValue(attrType uint32, value int64) bool {
	if !attrdef.IsFightAttr(attrType) {
		return false
	}
	idx := attrType - attrdef.FightAttrBegin
	changed := calc.values[idx] != value
	calc.values[idx] = value
	return changed
}

// GetValue 获取属性值
func (calc *FightAttrCalc) GetValue(attrType uint32) int64 {
	if !attrdef.IsFightAttr(attrType) {
		return 0
	}
	return calc.values[attrType-attrdef.FightAttrBegin]
}

// AddValue 属性值累加
func (calc *FightAttrCalc) AddValue(attrType uint32, delta int64) {
	if !attrdef.IsFightAttr(attrType) {
		return
	}
	idx := attrType - attrdef.FightAttrBegin
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
func (calc *FightAttrCalc) DoRange(cb func(attrType uint32, value int64)) {
	for idx, v := range calc.values {
		if v == 0 {
			continue
		}
		cb(uint32(idx)+attrdef.FightAttrBegin, v)
	}
}
