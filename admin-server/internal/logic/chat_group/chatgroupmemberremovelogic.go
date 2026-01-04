// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_group

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatGroupMemberRemoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatGroupMemberRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatGroupMemberRemoveLogic {
	return &ChatGroupMemberRemoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatGroupMemberRemoveLogic) ChatGroupMemberRemove(req *types.ChatGroupMemberRemoveReq) (resp *types.Response, err error) {
	// 获取当前用户
	_, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)
	chatUserRepo := repository.NewChatUserRepository(l.svcCtx.Repository)

	// 查询群组
	chat, err := chatRepo.FindByID(l.ctx, req.ChatId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeNotFound, "群组不存在", err)
	}

	// 验证是否为群组
	if chat.Type != 2 {
		return nil, errs.New(errs.CodeBadRequest, "该聊天不是群组")
	}

	// 验证是否已删除
	if chat.DeletedAt != 0 {
		return nil, errs.New(errs.CodeNotFound, "群组已删除")
	}

	// 检查用户是否在群组中
	chatUsers, err := chatUserRepo.FindByChatID(l.ctx, req.ChatId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询群组成员失败", err)
	}

	userInGroup := false
	for _, cu := range chatUsers {
		if cu.UserId == req.UserId {
			userInGroup = true
			break
		}
	}

	if !userInGroup {
		return nil, errs.New(errs.CodeBadRequest, "用户不在群组中")
	}

	// 移除成员
	err = chatUserRepo.DeleteByChatIDAndUserID(l.ctx, req.ChatId, req.UserId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "移除成员失败", err)
	}

	return &types.Response{
		Code:    0,
		Message: "移除成员成功",
	}, nil
}
