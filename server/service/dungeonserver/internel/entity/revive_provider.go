package entity

import "postapocgame/server/service/dungeonserver/internel/iface"

var reviveSceneProvider func() iface.IScene

// SetReviveSceneProvider 注册复活场景提供函数
func SetReviveSceneProvider(fn func() iface.IScene) {
	reviveSceneProvider = fn
}

func getReviveScene() iface.IScene {
	if reviveSceneProvider == nil {
		return nil
	}
	return reviveSceneProvider()
}

// GetReviveScene 对外暴露复活场景
func GetReviveScene() iface.IScene {
	return getReviveScene()
}
