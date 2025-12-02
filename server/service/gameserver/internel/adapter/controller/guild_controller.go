package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/di"
	guildusecase "postapocgame/server/service/gameserver/internel/usecase/guild"
)

// GuildController 公会系统控制器
type GuildController struct {
	createUC  *guildusecase.CreateGuildUseCase
	joinUC    *guildusecase.JoinGuildUseCase
	leaveUC   *guildusecase.LeaveGuildUseCase
	queryUC   *guildusecase.QueryGuildInfoUseCase
	presenter *presenter.GuildPresenter
}

// NewGuildController 创建控制器
func NewGuildController() *GuildController {
	container := di.GetContainer()
	playerRepo := container.PlayerGateway()
	publicActor := container.PublicActorGateway()
	return &GuildController{
		createUC:  guildusecase.NewCreateGuildUseCase(playerRepo, publicActor),
		joinUC:    guildusecase.NewJoinGuildUseCase(playerRepo, publicActor),
		leaveUC:   guildusecase.NewLeaveGuildUseCase(playerRepo, publicActor),
		queryUC:   guildusecase.NewQueryGuildInfoUseCase(playerRepo),
		presenter: presenter.NewGuildPresenter(container.NetworkGateway()),
	}
}

// HandleCreateGuild 创建公会
func (c *GuildController) HandleCreateGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil || playerRole.GetRoleInfo() == nil {
		return c.presenter.SendError(ctx, sessionID, "角色信息不存在")
	}
	var req protocol.C2SCreateGuildReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	if err := c.createUC.Execute(ctx, roleID, playerRole.GetRoleInfo().RoleName, req.GuildName); err != nil {
		return c.presenter.SendCreateResult(ctx, sessionID, false, err.Error())
	}
	return c.presenter.SendCreateResult(ctx, sessionID, true, "创建公会请求已提交")
}

// HandleJoinGuild 加入公会
func (c *GuildController) HandleJoinGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil || playerRole.GetRoleInfo() == nil {
		return c.presenter.SendError(ctx, sessionID, "角色信息不存在")
	}
	var req protocol.C2SJoinGuildReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	if err := c.joinUC.Execute(ctx, roleID, playerRole.GetRoleInfo().RoleName, req.GuildId); err != nil {
		return c.presenter.SendJoinResult(ctx, sessionID, false, err.Error())
	}
	return c.presenter.SendJoinResult(ctx, sessionID, true, "加入公会请求已提交")
}

// HandleLeaveGuild 离开公会
func (c *GuildController) HandleLeaveGuild(ctx context.Context, _ *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := c.leaveUC.Execute(ctx, roleID); err != nil {
		return c.presenter.SendLeaveResult(ctx, sessionID, false, err.Error())
	}
	return c.presenter.SendLeaveResult(ctx, sessionID, true, "离开公会成功")
}

// HandleQueryGuildInfo 查询公会信息
func (c *GuildController) HandleQueryGuildInfo(ctx context.Context, _ *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	data, execErr := c.queryUC.Execute(ctx, roleID)
	if execErr != nil {
		return c.presenter.SendError(ctx, sessionID, execErr.Error())
	}
	return c.presenter.SendGuildInfo(ctx, sessionID, data)
}
