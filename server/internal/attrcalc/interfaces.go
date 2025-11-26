package attrcalc

import (
	"context"

	"postapocgame/server/internal/protocol"
)

// SysAttrCalculator 各系统基础属性计算器
type SysAttrCalculator interface {
	CalculateAttrs(ctx context.Context) []*protocol.AttrSt
}

// SysAddRateCalculator 各系统加成属性计算器
type SysAddRateCalculator interface {
	CalculateAddRate(ctx context.Context, totalCalc *FightAttrCalc) []*protocol.AttrSt
}
