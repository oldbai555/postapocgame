package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/chat"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// ChatController 聊天控制器
type ChatController struct {
	worldUC   *chat.WorldChatUseCase
	privateUC *chat.PrivateChatUseCase
	presenter *presenter.ChatPresenter
}

func NewChatController() *ChatController {
	config := deps.ConfigGateway()
	publicActor := deps.PublicActorGateway()
	return &ChatController{
		worldUC:   chat.NewWorldChatUseCase(config, publicActor),
		privateUC: chat.NewPrivateChatUseCase(config, publicActor),
		presenter: presenter.NewChatPresenter(deps.NetworkGateway()),
	}
}

// HandleWorldChat 处理世界聊天
func (c *ChatController) HandleWorldChat(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil || playerRole.GetRoleInfo() == nil {
		return c.presenter.SendError(ctx, sessionID, "角色信息不存在")
	}
	var req protocol.C2SChatWorldReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	// 检查系统是否开启
	chatSys := system.GetChatSys(ctx)
	if chatSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "聊天系统未开启")
	}
	if err := c.worldUC.Execute(ctx, chatSys, roleID, playerRole.GetRoleInfo().RoleName, req.Content); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}

// HandlePrivateChat 处理私聊
func (c *ChatController) HandlePrivateChat(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
	if playerRole == nil || playerRole.GetRoleInfo() == nil {
		return c.presenter.SendError(ctx, sessionID, "角色信息不存在")
	}
	var req protocol.C2SChatPrivateReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	// 检查系统是否开启
	chatSys := system.GetChatSys(ctx)
	if chatSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "聊天系统未开启")
	}
	if err := c.privateUC.Execute(ctx, chatSys, roleID, playerRole.GetRoleInfo().RoleName, req.TargetId, req.Content); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}
