package fuben

import (
	"postapocgame/server/service/gameserver/internel/dungeonactor/iface"
)

var (
	defaultFuBen iface.IFuBen
)

// SetDefaultFuBen 设置默认副本
func SetDefaultFuBen(fb iface.IFuBen) {
	defaultFuBen = fb
}

// GetDefaultFuBen 获取默认副本
func GetDefaultFuBen() iface.IFuBen {
	return defaultFuBen
}
