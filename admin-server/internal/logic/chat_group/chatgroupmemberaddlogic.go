// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_group

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatGroupMemberAddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatGroupMemberAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatGroupMemberAddLogic {
	return &ChatGroupMemberAddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatGroupMemberAddLogic) ChatGroupMemberAdd(req *types.ChatGroupMemberAddReq) (resp *types.Response, err error) {
	// 获取当前用户
	_, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	if len(req.UserIds) == 0 {
		return nil, errs.New(errs.CodeBadRequest, "用户ID列表不能为空")
	}

	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)
	chatUserRepo := repository.NewChatUserRepository(l.svcCtx.Repository)
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)

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

	// 查询现有成员
	existingUsers, err := chatUserRepo.FindByChatID(l.ctx, req.ChatId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询群组成员失败", err)
	}

	existingUserMap := make(map[uint64]bool)
	for _, cu := range existingUsers {
		existingUserMap[cu.UserId] = true
	}

	now := time.Now().Unix()
	addedCount := 0

	// 添加成员
	for _, userId := range req.UserIds {
		// 检查是否已在群组中
		if existingUserMap[userId] {
			continue
		}

		// 验证用户是否存在且未删除
		user, err := userRepo.FindByID(l.ctx, userId)
		if err != nil {
			logx.Errorf("查询用户失败: userId=%d, err=%v", userId, err)
			continue
		}
		if user.DeletedAt != 0 {
			logx.Errorf("用户已删除: userId=%d", userId)
			continue
		}

		// 添加到群组
		chatUser := &model.ChatUser{
			ChatId:    req.ChatId,
			UserId:    userId,
			JoinedAt:  now,
			CreatedAt: now,
			UpdatedAt: now,
		}
		err = chatUserRepo.Create(l.ctx, chatUser)
		if err != nil {
			logx.Errorf("添加成员到群组失败: userId=%d, err=%v", userId, err)
			continue
		}

		addedCount++
		existingUserMap[userId] = true // 标记为已添加
	}

	if addedCount == 0 {
		return nil, errs.New(errs.CodeBadRequest, "没有可添加的成员（可能已全部在群组中或用户不存在）")
	}

	return &types.Response{
		Code:    0,
		Message: "添加成员成功",
	}, nil
}
