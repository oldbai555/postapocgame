package clientprotocol

import (
	"context"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
)

// ProtoTbl 协议注册表
var (
	ProtoTbl = make(map[uint16]Func)
)

type Func func(ctx context.Context, msg *network.ClientMessage) error

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
