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
	friendusecase "postapocgame/server/service/gameserver/internel/usecase/friend"
)

// FriendController 好友系统控制器
type FriendController struct {
	sendRequestUC  *friendusecase.SendFriendRequestUseCase
	respondReqUC   *friendusecase.RespondFriendRequestUseCase
	removeFriendUC *friendusecase.RemoveFriendUseCase
	queryListUC    *friendusecase.QueryFriendListUseCase
	blacklistUC    *friendusecase.BlacklistUseCase
	presenter      *presenter.FriendPresenter
}

// NewFriendController 创建控制器
func NewFriendController() *FriendController {
	container := di.GetContainer()
	playerRepo := container.PlayerGateway()
	publicActor := container.PublicActorGateway()
	blacklistRepo := container.BlacklistRepository()

	return &FriendController{
		sendRequestUC:  friendusecase.NewSendFriendRequestUseCase(playerRepo, publicActor, blacklistRepo),
		respondReqUC:   friendusecase.NewRespondFriendRequestUseCase(playerRepo, publicActor),
		removeFriendUC: friendusecase.NewRemoveFriendUseCase(playerRepo),
		queryListUC:    friendusecase.NewQueryFriendListUseCase(playerRepo, publicActor),
		blacklistUC:    friendusecase.NewBlacklistUseCase(blacklistRepo),
		presenter:      presenter.NewFriendPresenter(container.NetworkGateway()),
	}
}

// HandleAddFriend 处理添加好友协议
func (c *FriendController) HandleAddFriend(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)

	var req protocol.C2SAddFriendReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleName := ""
	if playerRole != nil && playerRole.GetRoleInfo() != nil {
		roleName = playerRole.GetRoleInfo().RoleName
	}

	if err := c.sendRequestUC.Execute(ctx, roleID, roleName, req.TargetId); err != nil {
		return c.presenter.SendAddFriendResult(ctx, sessionID, false, err.Error(), req.TargetId)
	}
	return c.presenter.SendAddFriendResult(ctx, sessionID, true, "好友申请已发送", req.TargetId)
}

// HandleRespondFriendReq 处理响应好友申请
func (c *FriendController) HandleRespondFriendReq(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SRespondFriendReqReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	accepted, message, execErr := c.respondReqUC.Execute(ctx, roleID, req.RequesterId, req.Accepted)
	if execErr != nil {
		return c.presenter.SendRespondResult(ctx, sessionID, false, execErr.Error(), req.RequesterId, req.Accepted)
	}
	return c.presenter.SendRespondResult(ctx, sessionID, true, message, req.RequesterId, accepted)
}

// HandleQueryFriendList 处理好友列表查询
func (c *FriendController) HandleQueryFriendList(ctx context.Context, _ *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	if err := c.queryListUC.Execute(ctx, roleID, sessionID); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}

// HandleRemoveFriend 处理删除好友
func (c *FriendController) HandleRemoveFriend(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SRemoveFriendReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	success, message, execErr := c.removeFriendUC.Execute(ctx, roleID, req.FriendId)
	if execErr != nil {
		return c.presenter.SendRemoveResult(ctx, sessionID, false, execErr.Error(), req.FriendId)
	}
	return c.presenter.SendRemoveResult(ctx, sessionID, success, message, req.FriendId)
}

// HandleAddToBlacklist 添加黑名单
func (c *FriendController) HandleAddToBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SAddToBlacklistReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	if err := c.blacklistUC.Add(ctx, roleID, req.TargetId, req.Reason); err != nil {
		return c.presenter.SendAddBlacklistResult(ctx, sessionID, false, err.Error())
	}
	return c.presenter.SendAddBlacklistResult(ctx, sessionID, true, "已添加到黑名单")
}

// HandleRemoveFromBlacklist 移除黑名单
func (c *FriendController) HandleRemoveFromBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SRemoveFromBlacklistReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	if err := c.blacklistUC.Remove(ctx, roleID, req.TargetId); err != nil {
		return c.presenter.SendRemoveBlacklistResult(ctx, sessionID, false, err.Error())
	}
	return c.presenter.SendRemoveBlacklistResult(ctx, sessionID, true, "已从黑名单移除")
}

// HandleQueryBlacklist 查询黑名单
func (c *FriendController) HandleQueryBlacklist(ctx context.Context, _ *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	list, execErr := c.blacklistUC.Query(ctx, roleID)
	if execErr != nil {
		return c.presenter.SendError(ctx, sessionID, execErr.Error())
	}
	return c.presenter.SendBlacklist(ctx, sessionID, list)
}
