package iface

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/protocol"
	"time"
)

// IMoveSys 移动系统接口
type IMoveSys interface {
	BindScene(scene IScene)
	UnbindScene(scene IScene)
	RunOne(now time.Time)

	HandleStartMove(scene IScene, req *protocol.C2SStartMoveReq) error
	HandleUpdateMove(scene IScene, req *protocol.C2SUpdateMoveReq) error
	HandleEndMove(scene IScene, req *protocol.C2SEndMoveReq) error

	MoveTo(pos *argsdef.Position, speed float64)
	Stop()
	ResetState()
}
