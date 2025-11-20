package clientprotocol

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/entitysystem"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

func getSceneByEntity(entity iface.IEntity) (iface.IScene, error) {
	if entity == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "entity missing")
	}
	scene, ok := entitymgr.GetEntityMgr().GetSceneByHandle(entity.GetHdl())
	if !ok {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "scene not bound")
	}
	return scene, nil
}

func broadcastSceneMessage(scene iface.IScene, protoId uint16, payload proto.Message) {
	entitysystem.BroadcastSceneProto(scene, protoId, payload)
}

func sendStopToEntity(entity iface.IEntity, seq uint32) {
	if entity == nil {
		return
	}
	pos := entity.GetPosition()
	_ = entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEntityStopMove), &protocol.S2CEntityStopMoveReq{
		EntityHdl: entity.GetHdl(),
		PosX:      pos.X,
		PosY:      pos.Y,
		Seq:       seq,
	})
}
