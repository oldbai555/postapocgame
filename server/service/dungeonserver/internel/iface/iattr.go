/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 属性系统接口
**/

package iface

import (
	"postapocgame/server/internal/attrdef"
)

// IAttrSys 属性系统接口
type IAttrSys interface {
	// GetAttrValue 获取属性值
	GetAttrValue(attrType attrdef.AttrType) attrdef.AttrValue

	// SetAttrValue 设置属性值
	SetAttrValue(attrType attrdef.AttrType, value attrdef.AttrValue)

	// AddAttrValue 增加属性值
	AddAttrValue(attrType attrdef.AttrType, delta attrdef.AttrValue)

	// GetAllCombatAttrs 获取所有战斗属性
	GetAllCombatAttrs() map[attrdef.AttrType]attrdef.AttrValue

	// GetAllExtraAttrs 获取所有非战斗属性
	GetAllExtraAttrs() map[attrdef.AttrType]attrdef.AttrValue

	// ResetCombatAttrs 重置战斗属性
	ResetCombatAttrs()

	// ResetExtraAttrs 重置非战斗属性
	ResetExtraAttrs()

	// ResetAll 重置所有属性
	ResetAll()

	// BatchSetAttrs 批量设置属性
	BatchSetAttrs(attrs map[attrdef.AttrType]attrdef.AttrValue)

	// BatchAddAttrs 批量增加属性
	BatchAddAttrs(attrs map[attrdef.AttrType]attrdef.AttrValue)
}
