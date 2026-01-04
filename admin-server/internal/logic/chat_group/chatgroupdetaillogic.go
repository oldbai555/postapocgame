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

type ChatGroupDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatGroupDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatGroupDetailLogic {
	return &ChatGroupDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatGroupDetailLogic) ChatGroupDetail(req *types.ChatGroupDetailReq) (resp *types.ChatGroupDetailResp, err error) {
	// 获取当前用户
	_, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)
	chatUserRepo := repository.NewChatUserRepository(l.svcCtx.Repository)

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

	// 查询成员列表
	chatUsers, err := chatUserRepo.FindByChatID(l.ctx, chat.Id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询群组成员失败", err)
	}

	// 获取成员详细信息
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	deptRepo := repository.NewDepartmentRepository(l.svcCtx.Repository)
	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	userRoleRepo := repository.NewUserRoleRepository(l.svcCtx.Repository)

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

	members := make([]types.ChatGroupMemberItem, 0, len(chatUsers))
	for _, cu := range chatUsers {
		user, err := userRepo.FindByID(l.ctx, cu.UserId)
		if err != nil || user.DeletedAt != 0 {
			continue // 跳过已删除的用户
		}

		member := types.ChatGroupMemberItem{
			UserId:   user.Id,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			JoinedAt: cu.JoinedAt,
		}

		// 获取部门名称
		if user.DepartmentId > 0 {
			if deptName, ok := deptMap[user.DepartmentId]; ok {
				member.DepartmentName = deptName
			}
		}

		// 获取角色名称列表
		roleIDs, _ := userRoleRepo.ListRoleIDsByUserID(l.ctx, user.Id)
		roleNames := make([]string, 0, len(roleIDs))
		for _, roleID := range roleIDs {
			if roleName, ok := roleMap[roleID]; ok {
				roleNames = append(roleNames, roleName)
			}
		}
		member.RoleNames = roleNames

		members = append(members, member)
	}

	resp = &types.ChatGroupDetailResp{
		Id:          chat.Id,
		Name:        chat.Name,
		Avatar:      chat.Avatar,
		Description: chat.Description,
		CreatedBy:   chat.CreatedBy,
		CreatedAt:   chat.CreatedAt,
		MemberCount: int64(len(members)),
		Members:     members,
	}

	return resp, nil
}
