package attrcalc

// AttrSet 管理各系统的增量/减量属性集合
type AttrSet struct {
	inc map[uint32]*FightAttrCalc
	dec map[uint32]*FightAttrCalc
}

// NewAttrSet 创建属性集合
func NewAttrSet() *AttrSet {
	return &AttrSet{
		inc: make(map[uint32]*FightAttrCalc),
		dec: make(map[uint32]*FightAttrCalc),
	}
}

// Reset 重置所有属性
func (set *AttrSet) Reset() {
	for _, calc := range set.inc {
		calc.Reset()
	}
	for _, calc := range set.dec {
		calc.Reset()
	}
}

// GetIncAttr 获取增量属性计算器
func (set *AttrSet) GetIncAttr(sysID uint32, create bool, autoReset bool) *FightAttrCalc {
	return set.getCalc(set.inc, sysID, create, autoReset)
}

// GetDecAttr 获取减量属性计算器
func (set *AttrSet) GetDecAttr(sysID uint32, create bool, autoReset bool) *FightAttrCalc {
	return set.getCalc(set.dec, sysID, create, autoReset)
}

func (set *AttrSet) getCalc(bucket map[uint32]*FightAttrCalc, sysID uint32, create bool, autoReset bool) *FightAttrCalc {
	calc, ok := bucket[sysID]
	if !ok {
		if !create {
			return nil
		}
		calc = NewFightAttrCalc()
		bucket[sysID] = calc
	} else if autoReset {
		calc.Reset()
	}
	return calc
}

// ResetProperty 将所有增减属性累计到目标计算器
func (set *AttrSet) ResetProperty(target *FightAttrCalc) {
	if target == nil {
		return
	}
	for _, calc := range set.inc {
		target.AddCalc(calc)
	}
	for _, calc := range set.dec {
		target.SubCalc(calc)
	}
}

// PackToMap 将当前增量属性导出为 map 结构
func (set *AttrSet) PackToMap() map[uint32]map[uint32]int64 {
	result := make(map[uint32]map[uint32]int64)
	for sysID, calc := range set.inc {
		attrMap := make(map[uint32]int64)
		calc.DoRange(func(attrType uint32, value int64) {
			attrMap[attrType] = value
		})
		if len(attrMap) > 0 {
			result[sysID] = attrMap
		}
	}
	return result
}
