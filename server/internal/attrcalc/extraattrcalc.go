/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package attrcalc

import (
	"postapocgame/server/internal/attrdef"
)

// ExtraAttrCalc 非战斗属性计算器
type ExtraAttrCalc struct {
	extraAttr [ExtraAttrCount]int64
}

// NewExtraAttrCalc 创建非战斗属性计算器
func NewExtraAttrCalc() *ExtraAttrCalc {
	return &ExtraAttrCalc{}
}

// Reset 重置所有属性
func (calc *ExtraAttrCalc) Reset() {
	copy(calc.extraAttr[:], ZeroExtraAttrs[:])
}

// SetValue 设置属性值
func (calc *ExtraAttrCalc) SetValue(attrType uint32, value int64) {
	if !attrdef.IsExtraAttr(attrType) {
		return
	}
	calc.extraAttr[attrType-attrdef.ExtraAttrBegin] = value
	return
}

// GetValue 获取属性值
func (calc *ExtraAttrCalc) GetValue(attrType uint32) int64 {
	if !attrdef.IsExtraAttr(attrType) {
		return 0
	}
	return calc.extraAttr[attrType-attrdef.ExtraAttrBegin]
}

// AddValue 增加属性值
func (calc *ExtraAttrCalc) AddValue(attrType uint32, delta int64) {
	if !attrdef.IsExtraAttr(attrType) {
		return
	}
	calc.extraAttr[attrType-attrdef.ExtraAttrBegin] += delta
	return
}

// DoRange 遍历所有属性
func (calc *ExtraAttrCalc) DoRange(cb func(attrType uint32, value int64)) {
	for i, v := range calc.extraAttr {
		if v != 0 {
			cb(uint32(i)+attrdef.ExtraAttrBegin, v)
		}
	}
}

// GetAll 获取所有属性
func (calc *ExtraAttrCalc) GetAll() map[uint32]int64 {
	result := make(map[uint32]int64)
	calc.DoRange(func(attrType uint32, value int64) {
		result[attrType] = value
	})
	return result
}
