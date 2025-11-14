package clientprotocol

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/dungeonserver/internel/entity"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

func init() {
	Register(uint16(protocol.C2SProtocol_C2SRevive), handleRevive)
}

func handleRevive(role iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SReviveReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 检查实体是否为角色实体
	roleEntity, ok := role.(*entity.RoleEntity)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "entity is not a role entity")
	}

	// 检查角色是否死亡
	if !roleEntity.IsDead() {
		resp := &protocol.S2CReviveResultReq{
			Success: false,
			Message: "角色未死亡，无需复活",
		}
		return roleEntity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CReviveResult), resp)
	}

	// 执行复活
	err := roleEntity.Revive()
	if err != nil {
		resp := &protocol.S2CReviveResultReq{
			Success: false,
			Message: err.Error(),
		}
		return roleEntity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CReviveResult), resp)
	}

	// Revive方法内部已经发送了响应，这里不需要再发送
	return nil
}
