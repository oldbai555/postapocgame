package entityhelper

import (
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

var defaultVisibleAttrTypes = []attrdef.AttrType{
	attrdef.AttrHP,
	attrdef.AttrMaxHP,
	attrdef.AttrMP,
	attrdef.AttrMaxMP,
}

// BuildAttrMap 构建实体属性快照（用于协议下发）
func BuildAttrMap(entity iface.IEntity, attrTypes ...attrdef.AttrType) map[uint32]int64 {
	if entity == nil {
		return nil
	}
	attrSys := entity.GetAttrSys()
	if attrSys == nil {
		return nil
	}
	if len(attrTypes) == 0 {
		attrTypes = DefaultVisibleAttrTypes()
	}

	attrs := make(map[uint32]int64, len(attrTypes))
	for _, attrType := range attrTypes {
		value := attrSys.GetAttrValue(attrType)
		attrs[uint32(attrType)] = int64(value)
	}
	return attrs
}

// DefaultVisibleAttrTypes 返回默认需要暴露给客户端的属性类型
func DefaultVisibleAttrTypes() []attrdef.AttrType {
	cp := make([]attrdef.AttrType, len(defaultVisibleAttrTypes))
	copy(cp, defaultVisibleAttrTypes)
	return cp
}
