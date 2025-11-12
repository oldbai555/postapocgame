/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

type IFuBen interface {
	Close()
	GetScene(sceneId uint32) IScene
	GetAllScenes() []IScene
	GetFbId() uint32
	IsExpired() bool
	GetFbType() uint32
	GetState() uint32
	GetPlayerCount() int
}
