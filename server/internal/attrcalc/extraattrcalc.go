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
	extraAttr [ExtraAttrCount]attrdef.AttrValue
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
func (calc *ExtraAttrCalc) SetValue(attrType attrdef.AttrType, value attrdef.AttrValue) {
	if !attrdef.IsExtraAttr(attrType) {
		return
	}
	calc.extraAttr[attrType-attrdef.ExtraAttrBegin] = value
	return
}

// GetValue 获取属性值
func (calc *ExtraAttrCalc) GetValue(attrType attrdef.AttrType) attrdef.AttrValue {
	if !attrdef.IsExtraAttr(attrType) {
		return 0
	}
	return calc.extraAttr[attrType-attrdef.ExtraAttrBegin]
}

// AddValue 增加属性值
func (calc *ExtraAttrCalc) AddValue(attrType attrdef.AttrType, delta attrdef.AttrValue) {
	if !attrdef.IsExtraAttr(attrType) {
		return
	}
	calc.extraAttr[attrType-attrdef.ExtraAttrBegin] += delta
	return
}

// DoRange 遍历所有属性
func (calc *ExtraAttrCalc) DoRange(cb func(attrType attrdef.AttrType, value attrdef.AttrValue)) {
	for i, v := range calc.extraAttr {
		if v != 0 {
			cb(attrdef.AttrType(i)+attrdef.ExtraAttrBegin, v)
		}
	}
}

// GetAll 获取所有属性
func (calc *ExtraAttrCalc) GetAll() map[attrdef.AttrType]attrdef.AttrValue {
	result := make(map[attrdef.AttrType]attrdef.AttrValue)
	calc.DoRange(func(attrType attrdef.AttrType, value attrdef.AttrValue) {
		result[attrType] = value
	})
	return result
}
