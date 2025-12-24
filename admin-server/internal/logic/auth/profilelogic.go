// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileLogic) Profile() (resp *types.ProfileResp, err error) {
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 获取用户权限
	userRoleRepo := repository.NewUserRoleRepository(l.svcCtx.Repository)
	roleIDs, err := userRoleRepo.ListRoleIDsByUserID(l.ctx, user.UserID)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "获取用户角色失败", err)
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	perms, err := permissionRepo.ListByRoleIDs(l.ctx, roleIDs)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "获取用户权限失败", err)
	}

	codes := make([]string, 0, len(perms))
	seen := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		if _, ok := seen[p.Code]; ok {
			continue
		}
		seen[p.Code] = struct{}{}
		codes = append(codes, p.Code)
	}

	return &types.ProfileResp{
		Id:          user.UserID,
		Username:    user.Username,
		Permissions: codes,
	}, nil
}
