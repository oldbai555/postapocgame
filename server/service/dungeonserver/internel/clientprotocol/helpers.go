package clientprotocol

import (
	"postapocgame/server/internal"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
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

func broadcastSceneMessage(scene iface.IScene, protoId uint16, payload interface{}) {
	if scene == nil || payload == nil {
		return
	}
	data, err := internal.Marshal(payload)
	if err != nil {
		log.Errorf("broadcast marshal failed: %v", err)
		return
	}
	for _, et := range scene.GetAllEntities() {
		_ = et.SendMessage(protoId, data)
	}
}

func sendStopToEntity(entity iface.IEntity, seq uint32) {
	if entity == nil {
		return
	}
	pos := entity.GetPosition()
	_ = entity.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEntityStopMove), &protocol.S2CEntityStopMoveReq{
		EntityHdl: entity.GetHdl(),
		PosX:      pos.X,
		PosY:      pos.Y,
		Seq:       seq,
	})
}
