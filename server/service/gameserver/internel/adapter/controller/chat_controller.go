package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/adapter/system"
	"postapocgame/server/service/gameserver/internel/di"
	chatusescase "postapocgame/server/service/gameserver/internel/usecase/chat"
)

// ChatController 聊天控制器
type ChatController struct {
	worldUC   *chatusescase.WorldChatUseCase
	privateUC *chatusescase.PrivateChatUseCase
	presenter *presenter.ChatPresenter
}

func NewChatController() *ChatController {
	container := di.GetContainer()
	config := container.ConfigGateway()
	publicActor := container.PublicActorGateway()
	return &ChatController{
		worldUC:   chatusescase.NewWorldChatUseCase(config, publicActor),
		privateUC: chatusescase.NewPrivateChatUseCase(config, publicActor),
		presenter: presenter.NewChatPresenter(container.NetworkGateway()),
	}
}

// HandleWorldChat 处理世界聊天
func (c *ChatController) HandleWorldChat(ctx context.Context, msg *network.ClientMessage) error {
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
	var req protocol.C2SChatWorldReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	chatSys := system.GetChatSys(ctx)
	if chatSys == nil {
		return c.presenter.SendError(ctx, sessionID, "聊天系统未初始化")
	}
	if err := c.worldUC.Execute(ctx, chatSys, roleID, playerRole.GetRoleInfo().RoleName, req.Content); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}

// HandlePrivateChat 处理私聊
func (c *ChatController) HandlePrivateChat(ctx context.Context, msg *network.ClientMessage) error {
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
	var req protocol.C2SChatPrivateReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	chatSys := system.GetChatSys(ctx)
	if chatSys == nil {
		return c.presenter.SendError(ctx, sessionID, "聊天系统未初始化")
	}
	if err := c.privateUC.Execute(ctx, chatSys, roleID, playerRole.GetRoleInfo().RoleName, req.TargetId, req.Content); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}
