package iface

import (
	"postapocgame/server/internal/protocol"
)

// IMoveSys 移动系统接口
type IMoveSys interface {
	BindScene(scene IScene)
	UnbindScene(scene IScene)

	HandleStartMove(scene IScene, req *protocol.C2SStartMoveReq) error
	HandleUpdateMove(scene IScene, req *protocol.C2SUpdateMoveReq) error
	HandleEndMove(scene IScene, req *protocol.C2SEndMoveReq) error

	StopMove(broadcast, sendToSelf bool)
	ResetState()
	IsMoving() bool
}
