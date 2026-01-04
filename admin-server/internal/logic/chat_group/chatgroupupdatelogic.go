// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_group

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatGroupUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatGroupUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatGroupUpdateLogic {
	return &ChatGroupUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatGroupUpdateLogic) ChatGroupUpdate(req *types.ChatGroupUpdateReq) (resp *types.Response, err error) {
	// 获取当前用户
	_, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)

	// 查询群组
	chat, err := chatRepo.FindByID(l.ctx, req.Id)
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

	// 更新字段
	updated := false
	if req.Name != "" && req.Name != chat.Name {
		chat.Name = req.Name
		updated = true
	}
	if req.Avatar != "" && req.Avatar != chat.Avatar {
		chat.Avatar = req.Avatar
		updated = true
	}
	if req.Description != "" && req.Description != chat.Description {
		chat.Description = req.Description
		updated = true
	}

	if !updated {
		return &types.Response{
			Code:    0,
			Message: "无需更新",
		}, nil
	}

	chat.UpdatedAt = time.Now().Unix()
	err = chatRepo.Update(l.ctx, chat)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "更新群组失败", err)
	}

	return &types.Response{
		Code:    0,
		Message: "更新群组成功",
	}, nil
}
