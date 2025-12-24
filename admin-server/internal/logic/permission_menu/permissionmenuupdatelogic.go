// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission_menu

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionMenuUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermissionMenuUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PermissionMenuUpdateLogic {
	return &PermissionMenuUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermissionMenuUpdateLogic) PermissionMenuUpdate(req *types.PermissionMenuUpdateReq) error {
	if req.PermissionId == 0 {
		return errs.New(errs.CodeBadRequest, "权限ID不能为空")
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	// 验证权限是否存在
	_, err := permissionRepo.FindByID(l.ctx, req.PermissionId)
	if err != nil {
		return errs.Wrap(errs.CodeBadRequest, "权限不存在", err)
	}

	menuRepo := repository.NewMenuRepository(l.svcCtx.Repository)
	// 验证所有菜单是否存在
	for _, menuID := range req.MenuIds {
		_, err := menuRepo.FindByID(l.ctx, menuID)
		if err != nil {
			return errs.Wrap(errs.CodeBadRequest, "菜单不存在", err)
		}
	}

	permissionMenuRepo := repository.NewPermissionMenuRepository(l.svcCtx.Repository)
	if err := permissionMenuRepo.UpdatePermissionMenus(l.ctx, req.PermissionId, req.MenuIds); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新权限菜单失败", err)
	}
	return nil
}
