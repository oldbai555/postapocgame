package fuben

import (
	"postapocgame/server/service/dungeonserver/internel/entity"
	"postapocgame/server/service/dungeonserver/internel/iface"
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

func init() {
	entity.SetReviveSceneProvider(func() iface.IScene {
		if defaultFuBen == nil {
			return nil
		}
		return defaultFuBen.GetScene(1)
	})
}
