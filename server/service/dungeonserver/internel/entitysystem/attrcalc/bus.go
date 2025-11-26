package attrcalc

import (
	"sync"

	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

// IncAttrCalcFn 增量属性计算回调
type IncAttrCalcFn func(owner iface.IEntity, calc *icalc.FightAttrCalc)

// DecAttrCalcFn 减量属性计算回调
type DecAttrCalcFn func(owner iface.IEntity, calc *icalc.FightAttrCalc)

var (
	mu           sync.RWMutex
	incCallbacks = make(map[uint32]IncAttrCalcFn)
	decCallbacks = make(map[uint32]DecAttrCalcFn)
)

// RegIncAttrCalcFn 注册增量属性计算回调
func RegIncAttrCalcFn(sysId uint32, fn IncAttrCalcFn) {
	if fn == nil {
		log.Errorf("RegIncAttrCalcFn: fn is nil for sysId=%d", sysId)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := incCallbacks[sysId]; exists {
		log.Errorf("RegIncAttrCalcFn: sysId=%d already registered", sysId)
		return
	}
	incCallbacks[sysId] = fn
	log.Infof("RegIncAttrCalcFn: registered sysId=%d", sysId)
}

// RegDecAttrCalcFn 注册减量属性计算回调
func RegDecAttrCalcFn(sysId uint32, fn DecAttrCalcFn) {
	if fn == nil {
		log.Errorf("RegDecAttrCalcFn: fn is nil for sysId=%d", sysId)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := decCallbacks[sysId]; exists {
		log.Errorf("RegDecAttrCalcFn: sysId=%d already registered", sysId)
		return
	}
	decCallbacks[sysId] = fn
	log.Infof("RegDecAttrCalcFn: registered sysId=%d", sysId)
}

// GetIncAttrCalcFn 获取增量属性计算回调
func GetIncAttrCalcFn(sysId uint32) IncAttrCalcFn {
	mu.RLock()
	defer mu.RUnlock()
	return incCallbacks[sysId]
}

// GetDecAttrCalcFn 获取减量属性计算回调
func GetDecAttrCalcFn(sysId uint32) DecAttrCalcFn {
	mu.RLock()
	defer mu.RUnlock()
	return decCallbacks[sysId]
}
