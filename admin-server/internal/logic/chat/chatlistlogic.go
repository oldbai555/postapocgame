// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatListLogic {
	return &ChatListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatListLogic) ChatList(req *types.ChatListReq) (resp *types.ChatListResp, err error) {
	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 查询用户参与的所有聊天（通过chat_user关联表）
	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)
	chats, err := chatRepo.FindByUserID(l.ctx, user.UserID)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询聊天列表失败", err)
	}

	// 查询部门和角色信息（用于私聊显示）
	deptRepo := repository.NewDepartmentRepository(l.svcCtx.Repository)
	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	userRoleRepo := repository.NewUserRoleRepository(l.svcCtx.Repository)
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)

	// 构建部门ID到名称的映射
	deptMap := make(map[uint64]string)
	allDepts, _ := deptRepo.ListAll(l.ctx)
	for _, dept := range allDepts {
		if dept.DeletedAt == 0 {
			deptMap[dept.Id] = dept.Name
		}
	}

	// 构建角色ID到名称的映射
	roleMap := make(map[uint64]string)
	allRoles, _, _ := roleRepo.FindPage(l.ctx, 1, 10000, "")
	for _, role := range allRoles {
		if role.DeletedAt == 0 {
			roleMap[role.Id] = role.Name
		}
	}

	items := make([]types.ChatItem, 0, len(chats))
	for _, chat := range chats {
		item := types.ChatItem{
			ChatId:      chat.Id,
			Name:        chat.Name,
			ChatType:    int64(chat.Type),
			Avatar:      chat.Avatar,
			Description: chat.Description,
		}

		// 如果是私聊（type=1），需要获取对方用户信息
		if chat.Type == 1 {
			// 查询私聊中的另一个用户（不是当前用户的那个）
			chatUserRepo := repository.NewChatUserRepository(l.svcCtx.Repository)
			chatUsers, err := chatUserRepo.FindByChatID(l.ctx, chat.Id)
			if err == nil && len(chatUsers) == 2 {
				// 找到对方用户ID
				var otherUserID uint64
				for _, chatUser := range chatUsers {
					if chatUser.UserId != user.UserID {
						otherUserID = chatUser.UserId
						break
					}
				}

				if otherUserID > 0 {
					// 获取对方用户信息
					otherUser, err := userRepo.FindByID(l.ctx, otherUserID)
					if err == nil && otherUser.DeletedAt == 0 && otherUser.Status == 1 {
						item.UserId = otherUser.Id
						item.Username = otherUser.Username
						item.Nickname = otherUser.Nickname
						// 显示名称：优先使用昵称，否则使用用户名
						if otherUser.Nickname != "" {
							item.Name = otherUser.Nickname
						} else {
							item.Name = otherUser.Username
						}
						item.Avatar = otherUser.Avatar

						// 获取部门名称
						if otherUser.DepartmentId > 0 {
							if deptName, ok := deptMap[otherUser.DepartmentId]; ok {
								item.DepartmentName = deptName
							}
						}

						// 获取角色名称列表
						roleIDs, _ := userRoleRepo.ListRoleIDsByUserID(l.ctx, otherUser.Id)
						roleNames := make([]string, 0, len(roleIDs))
						for _, roleID := range roleIDs {
							if roleName, ok := roleMap[roleID]; ok {
								roleNames = append(roleNames, roleName)
							}
						}
						item.RoleNames = roleNames
					}
				}
			}
		}

		// TODO: 查询未读消息数和最后一条消息（后续实现）
		item.UnreadCount = 0
		item.LastMessage = ""
		item.LastMessageAt = 0

		items = append(items, item)
	}

	return &types.ChatListResp{
		List: items,
	}, nil
}
