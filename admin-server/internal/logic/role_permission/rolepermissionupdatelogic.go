// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role_permission

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type RolePermissionUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRolePermissionUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePermissionUpdateLogic {
	return &RolePermissionUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RolePermissionUpdateLogic) RolePermissionUpdate(req *types.RolePermissionUpdateReq) error {
	if req.RoleId == 0 {
		return errs.New(errs.CodeBadRequest, "角色ID不能为空")
	}

	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	// 验证角色是否存在
	_, err := roleRepo.FindByID(l.ctx, req.RoleId)
	if err != nil {
		return errs.Wrap(errs.CodeBadRequest, "角色不存在", err)
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	// 验证所有权限是否存在
	for _, permID := range req.PermissionIds {
		_, err := permissionRepo.FindByID(l.ctx, permID)
		if err != nil {
			return errs.Wrap(errs.CodeBadRequest, "权限不存在", err)
		}
	}

	rolePermissionRepo := repository.NewRolePermissionRepository(l.svcCtx.Repository)
	if err := rolePermissionRepo.UpdateRolePermissions(l.ctx, req.RoleId, req.PermissionIds); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新角色权限失败", err)
	}
	return nil
}
