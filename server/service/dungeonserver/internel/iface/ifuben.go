/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

import (
	"postapocgame/server/internal/custom_id"
)

type IFuBen interface {
	Close()
	GetScene(sceneId uint32) IScene
	GetAllScenes() []IScene
	GetFbId() uint32
	IsExpired() bool
	GetFbType() custom_id.FuBenType
	GetState() custom_id.FuBenState
	GetPlayerCount() int
}
