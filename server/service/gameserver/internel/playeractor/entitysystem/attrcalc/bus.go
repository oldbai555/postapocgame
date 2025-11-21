package attrcalc

import (
	"context"
	"sync"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
)

// Calculator 属性计算器接口
type Calculator interface {
	CalculateAttrs(ctx context.Context) []*protocol.AttrSt
}

// Provider 提供当前上下文的属性计算器
type Provider func(ctx context.Context) Calculator

// Bus 全局计算器总线（可克隆）
type Bus struct {
	mu        sync.RWMutex
	providers map[uint32]Provider
}

// NewBus 创建总线
func NewBus() *Bus {
	return &Bus{
		providers: make(map[uint32]Provider),
	}
}

// Register 注册计算器提供器
func (b *Bus) Register(saAttrSysId uint32, provider Provider) {
	if provider == nil {
		log.Warnf("attrcalc: provider is nil for SaAttrSys=%d", saAttrSysId)
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.providers[saAttrSysId] = provider
	log.Infof("attrcalc: registered provider for SaAttrSys=%d", saAttrSysId)
}

// CloneInstantiated 根据上下文克隆计算器集合
func (b *Bus) CloneInstantiated(ctx context.Context) map[uint32]Calculator {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if ctx == nil {
		ctx = context.Background()
	}

	result := make(map[uint32]Calculator, len(b.providers))
	for saAttrSysId, provider := range b.providers {
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

var defaultBus = NewBus()

// Register 注册全局计算器提供器
func Register(saAttrSysId uint32, provider Provider) {
	defaultBus.Register(saAttrSysId, provider)
}

// CloneCalculators 克隆当前上下文的计算器集合
func CloneCalculators(ctx context.Context) map[uint32]Calculator {
	return defaultBus.CloneInstantiated(ctx)
}
