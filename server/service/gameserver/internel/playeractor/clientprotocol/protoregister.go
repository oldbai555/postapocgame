package clientprotocol

import (
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

// 协议注册表
var (
	ProtoTbl = make(map[uint16]Func)
)

type Func func(actor iface.IPlayerRole, msg *network.ClientMessage) error

func Register(protoIdH, protoIdL uint16, f Func) {
	protoId := protoIdH<<8 | protoIdL
	if _, ok := ProtoTbl[protoId]; ok {
		log.Stackf("cmdId:%d %d register repeat.", protoIdH, protoIdL)
		return
	}
	ProtoTbl[protoId] = f
}
func RegisterProtoId(protoId uint16, f Func) {
	if _, ok := ProtoTbl[protoId]; ok {
		log.Stackf("cmdId:%d register repeat.", protoId)
		return
	}
	ProtoTbl[protoId] = f
}

func GetFunc(protoIdH, protoIdL uint16) Func {
	protoId := protoIdH<<8 | protoIdL
	f := ProtoTbl[protoId]
	if f == nil {
		return nil
	}
	return f
}
