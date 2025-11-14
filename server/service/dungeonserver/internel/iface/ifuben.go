/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

import (
	"postapocgame/server/internal/jsonconf"
	"time"
)

type IFuBen interface {
	Close()
	InitScenes(sceneConfigs []jsonconf.SceneConfig)
	SetDifficulty(difficulty uint32)
	OnPlayerEnter(sessionId string) error
	GetScene(sceneId uint32) IScene
	GetAllScenes() []IScene
	GetFbId() uint32
	IsExpired() bool
	GetFbType() uint32
	GetState() uint32
	GetPlayerCount() int
	RunOne(now time.Time)
}
