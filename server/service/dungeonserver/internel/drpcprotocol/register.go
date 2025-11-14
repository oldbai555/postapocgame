/**
 * @Author: zjj
 * @Date: 2025/11/14
 * @Desc:
**/

package drpcprotocol

import (
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/log"
)

// 协议注册表
var (
	ProtoTbl = make(map[uint16]Func)
)

type Func func(msg actor.IActorMessage) error

func Register(protoId uint16, f Func) {
	if _, ok := ProtoTbl[protoId]; ok {
		log.Stackf("cmdId:%d register repeat.", protoId)
		return
	}
	ProtoTbl[protoId] = f
}

func GetFunc(protoId uint16) Func {
	f := ProtoTbl[protoId]
	if f == nil {
		return nil
	}
	return f
}

// GetRegisteredProtocols 获取已注册的协议列表
func GetRegisteredProtocols() []uint16 {
	protocols := make([]uint16, 0, len(ProtoTbl))
	for protoId := range ProtoTbl {
		protocols = append(protocols, protoId)
	}
	return protocols
}
