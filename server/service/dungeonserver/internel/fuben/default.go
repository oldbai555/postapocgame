package fuben

import (
	"sync"

	"postapocgame/server/service/dungeonserver/internel/iface"
)

var (
	defaultFuBen iface.IFuBen
	defaultMu    sync.RWMutex
)

// SetDefaultFuBen 设置默认副本
func SetDefaultFuBen(fb iface.IFuBen) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultFuBen = fb
}

// GetDefaultFuBen 获取默认副本
func GetDefaultFuBen() iface.IFuBen {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultFuBen
}
