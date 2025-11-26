package attrcalc

import (
	"context"
	"sync"

	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/pkg/log"
)

// AddRateCalculator 属性加成计算器接口（统一定义）
type AddRateCalculator = icalc.SysAddRateCalculator

// AddRateProvider 提供属性加成计算器
type AddRateProvider func(ctx context.Context) AddRateCalculator

type addRateBus struct {
	mu        sync.RWMutex
	providers map[uint32]AddRateProvider
}

var defaultAddRateBus = &addRateBus{
	providers: make(map[uint32]AddRateProvider),
}

// RegisterAddRate 注册加成计算器
func RegisterAddRate(saAttrSysId uint32, provider AddRateProvider) {
	if provider == nil {
		log.Warnf("attrcalc: add-rate provider is nil for SaAttrSys=%d", saAttrSysId)
		return
	}
	defaultAddRateBus.mu.Lock()
	defer defaultAddRateBus.mu.Unlock()
	defaultAddRateBus.providers[saAttrSysId] = provider
	log.Infof("attrcalc: registered add-rate provider for SaAttrSys=%d", saAttrSysId)
}

// CloneAddRateCalculators 克隆加成计算器集合
func CloneAddRateCalculators(ctx context.Context) map[uint32]AddRateCalculator {
	defaultAddRateBus.mu.RLock()
	defer defaultAddRateBus.mu.RUnlock()

	if ctx == nil {
		ctx = context.Background()
	}

	result := make(map[uint32]AddRateCalculator, len(defaultAddRateBus.providers))
	for saAttrSysId, provider := range defaultAddRateBus.providers {
		if provider == nil {
			continue
		}
		calculator := provider(ctx)
		if calculator == nil {
			continue
		}
		result[saAttrSysId] = calculator
	}
	return result
}
