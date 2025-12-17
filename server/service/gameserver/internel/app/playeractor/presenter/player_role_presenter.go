package presenter

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/model"
	"postapocgame/server/service/gameserver/internel/app/playeractor/gateway"
	playerrole2 "postapocgame/server/service/gameserver/internel/app/playeractor/service/playerrole"

	"postapocgame/server/internal/protocol"
)

// PlayerRolePresenter 角色相关回包
type PlayerRolePresenter struct {
	network gateway.NetworkGateway
}

// NewPlayerRolePresenter 创建 Presenter
func NewPlayerRolePresenter(network gateway.NetworkGateway) *PlayerRolePresenter {
	return &PlayerRolePresenter{network: network}
}

// SendRoleList 发送角色列表
func (p *PlayerRolePresenter) SendRoleList(ctx context.Context, sessionID string, result *playerrole2.QueryRolesResult) error {
	resp := &protocol.S2CRoleListReq{
		RoleList: convertRoles(result.Roles),
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CRoleList), resp)
}

// SendCreateRoleResult 发送创建角色结果
func (p *PlayerRolePresenter) SendCreateRoleResult(ctx context.Context, sessionID string, result *playerrole2.CreateRoleResult) error {
	var roleData *protocol.PlayerSimpleData
	if result.Role != nil {
		roleData = toProtoRole(result.Role)
	}

	resp := &protocol.S2CCreateRoleResultReq{
		Job:      0,
		Sex:      0,
		RoleName: "",
	}
	if roleData != nil {
		resp.Job = roleData.Job
		resp.Sex = roleData.Sex
		resp.RoleName = roleData.RoleName
	}

	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CCreateRoleResult), resp)
}

func convertRoles(roles []*model.Role) []*protocol.PlayerSimpleData {
	if len(roles) == 0 {
		return []*protocol.PlayerSimpleData{}
	}
	result := make([]*protocol.PlayerSimpleData, 0, len(roles))
	for _, r := range roles {
		result = append(result, toProtoRole(r))
	}
	return result
}

func toProtoRole(role *model.Role) *protocol.PlayerSimpleData {
	if role == nil {
		return nil
	}
	return &protocol.PlayerSimpleData{
		RoleId:   role.ID,
		Job:      role.Job,
		Sex:      role.Sex,
		RoleName: role.RoleName,
		Level:    role.Level,
	}
}
