package controller

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/gateway"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/model"
	playerrole2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/playerrole"

	"google.golang.org/protobuf/proto"

	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// PlayerRoleController 负责角色相关协议
type PlayerRoleController struct {
	queryUseCase   *playerrole2.QueryRolesUseCase
	createUseCase  *playerrole2.CreateRoleUseCase
	presenter      *presenter.PlayerRolePresenter
	sessionGateway gateway.SessionGateway
}

// NewPlayerRoleController 创建控制器
func NewPlayerRoleController() *PlayerRoleController {
	return &PlayerRoleController{
		queryUseCase:   playerrole2.NewQueryRolesUseCase(deps.RoleRepository()),
		createUseCase:  playerrole2.NewCreateRoleUseCase(deps.RoleRepository()),
		presenter:      presenter.NewPlayerRolePresenter(deps.NetworkGateway()),
		sessionGateway: deps.SessionGateway(),
	}
}

// HandleQueryRoles 处理角色列表
func (c *PlayerRoleController) HandleQueryRoles(ctx context.Context, _ *network.ClientMessage) error {
	sessionID, err := getSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	session := c.sessionGateway.GetSession(sessionID)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	accountID := session.GetAccountID()
	if accountID == 0 {
		return c.presenter.SendRoleList(ctx, sessionID, &playerrole2.QueryRolesResult{
			Roles: []*model.Role{},
		})
	}

	result, err := c.queryUseCase.Execute(ctx, uint64(accountID))
	if err != nil {
		return err
	}

	return c.presenter.SendRoleList(ctx, sessionID, result)
}

// HandleCreateRole 处理创建角色
func (c *PlayerRoleController) HandleCreateRole(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := getSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	session := c.sessionGateway.GetSession(sessionID)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	var req protocol.C2SCreateRoleReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	accountID := session.GetAccountID()
	if accountID == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "account not logged in")
	}

	var roleName string
	var job, sex uint32
	if req.RoleData != nil {
		roleName = req.RoleData.RoleName
		job = req.RoleData.Job
		sex = req.RoleData.Sex
	}

	result, err := c.createUseCase.Execute(ctx, playerrole2.CreateRoleInput{
		AccountID: uint64(accountID),
		RoleName:  roleName,
		Job:       job,
		Sex:       sex,
	})
	if err != nil {
		return err
	}

	return c.presenter.SendCreateRoleResult(ctx, sessionID, result)
}
