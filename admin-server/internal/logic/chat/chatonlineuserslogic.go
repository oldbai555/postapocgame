// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatOnlineUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatOnlineUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatOnlineUsersLogic {
	return &ChatOnlineUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatOnlineUsersLogic) ChatOnlineUsers() (resp *types.ChatOnlineUserResp, err error) {
	onlineUserRepo := repository.NewChatOnlineUserRepository(l.svcCtx.Repository)
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)

	// 查询所有在线用户
	onlineUsers, err := onlineUserRepo.FindAll(l.ctx)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询在线用户失败", err)
	}

	// 去重用户ID，并查询用户信息
	userIDSet := make(map[uint64]struct{})
	items := make([]types.ChatOnlineUserItem, 0)
	for _, onlineUser := range onlineUsers {
		if _, exists := userIDSet[onlineUser.UserId]; exists {
			continue // 已处理过该用户
		}
		userIDSet[onlineUser.UserId] = struct{}{}

		// 查询用户信息
		user, err := userRepo.FindByID(l.ctx, onlineUser.UserId)
		if err != nil {
			continue // 用户不存在，跳过
		}

		items = append(items, types.ChatOnlineUserItem{
			UserId:   user.Id,
			UserName: user.Username,
			Avatar:   "", // 暂时为空，后续可以扩展
		})
	}

	return &types.ChatOnlineUserResp{
		List: items,
	}, nil
}
