// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_role

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRoleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRoleUpdateLogic {
	return &UserRoleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRoleUpdateLogic) UserRoleUpdate(req *types.UserRoleUpdateReq) error {
	if req.UserId == 0 {
		return errs.New(errs.CodeBadRequest, "用户ID不能为空")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	// 验证用户是否存在
	_, err := userRepo.FindByID(l.ctx, req.UserId)
	if err != nil {
		return errs.Wrap(errs.CodeBadRequest, "用户不存在", err)
	}

	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	// 验证所有角色是否存在，并检查是否包含超级管理员角色
	for _, roleID := range req.RoleIds {
		// 不允许分配超级管理员角色（id=1）
		if roleID == 1 {
			return errs.New(errs.CodeBadRequest, "不允许分配超级管理员角色")
		}

		_, err := roleRepo.FindByID(l.ctx, roleID)
		if err != nil {
			return errs.Wrap(errs.CodeBadRequest, "角色不存在", err)
		}
	}

	userRoleRepo := repository.NewUserRoleRepository(l.svcCtx.Repository)
	if err := userRoleRepo.UpdateUserRoles(l.ctx, req.UserId, req.RoleIds); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新用户角色失败", err)
	}
	return nil
}
