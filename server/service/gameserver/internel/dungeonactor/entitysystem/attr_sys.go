/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 属性系统实现
**/

package entitysystem

import (
	"postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/service/gameserver/internel/dungeonactor/iface"
)

var _ iface.IAttrSys = (*AttrSys)(nil)

// AttrSys 属性系统
type AttrSys struct {
	owner     iface.IEntity
	attrSet   *attrcalc.AttrSet
	fightAttr *attrcalc.FightAttrCalc
	extraAttr *attrcalc.ExtraAttrCalc
}

// NewAttrSys 创建属性系统
func NewAttrSys(owner iface.IEntity) *AttrSys {
	return &AttrSys{
		owner:     owner,
		attrSet:   attrcalc.NewAttrSet(),
		fightAttr: attrcalc.NewFightAttrCalc(),
		extraAttr: attrcalc.NewExtraAttrCalc(),
	}
}

// GetAttrValue 获取属性值
func (as *AttrSys) GetAttrValue(attrType uint32) int64 {
	if attrdef.IsFightAttr(attrType) {
		return as.fightAttr.GetValue(attrType)
	}
	if attrdef.IsExtraAttr(attrType) {
		return as.extraAttr.GetValue(attrType)
	}
	return 0
}

func (as *AttrSys) SetAttrValue(attrType uint32, value int64) {
	if attrdef.IsFightAttr(attrType) {
		as.fightAttr.SetValue(attrType, value)
		return
	}
	if attrdef.IsExtraAttr(attrType) {
		as.extraAttr.SetValue(attrType, value)
	}
}

func (as *AttrSys) AddAttrValue(attrType uint32, delta int64) {
	if attrdef.IsFightAttr(attrType) {
		as.fightAttr.AddValue(attrType, delta)
		return
	}
	if attrdef.IsExtraAttr(attrType) {
		as.extraAttr.AddValue(attrType, delta)
	}
}

// RunOne 每帧更新（由实体 RunOne 调用）
func (as *AttrSys) RunOne() {
}
