package clientprotocol

import (
	"postapocgame/server/internal"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

func init() {
	Register(uint16(protocol.C2SProtocol_C2SStartMove), handleStartMove)
	Register(uint16(protocol.C2SProtocol_C2SUpdateMove), handleUpdateMove)
	Register(uint16(protocol.C2SProtocol_C2SEndMove), handleEndMove)
}

func handleStartMove(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SStartMoveReq
	if err := internal.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}
	moveSys := entity.GetMoveSys()
	if moveSys == nil {
		return nil
	}
	return moveSys.HandleStartMove(scene, &req)
}

func handleUpdateMove(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SUpdateMoveReq
	if err := internal.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}
	moveSys := entity.GetMoveSys()
	if moveSys == nil {
		return nil
	}
	return moveSys.HandleUpdateMove(scene, &req)
}

func handleEndMove(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SEndMoveReq
	if err := internal.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}
	moveSys := entity.GetMoveSys()
	if moveSys == nil {
		return nil
	}
	return moveSys.HandleEndMove(scene, &req)
}
