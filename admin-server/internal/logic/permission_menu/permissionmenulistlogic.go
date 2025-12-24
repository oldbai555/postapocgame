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

type PermissionMenuListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermissionMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PermissionMenuListLogic {
	return &PermissionMenuListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermissionMenuListLogic) PermissionMenuList(req *types.PermissionMenuListReq) (resp *types.PermissionMenuListResp, err error) {
	if req.PermissionId == 0 {
		return nil, errs.New(errs.CodeBadRequest, "权限ID不能为空")
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	// 验证权限是否存在
	_, err = permissionRepo.FindByID(l.ctx, req.PermissionId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeBadRequest, "权限不存在", err)
	}

	permissionMenuRepo := repository.NewPermissionMenuRepository(l.svcCtx.Repository)
	menuIDs, err := permissionMenuRepo.ListMenuIDsByPermissionID(l.ctx, req.PermissionId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询权限菜单失败", err)
	}

	return &types.PermissionMenuListResp{
		MenuIds: menuIDs,
	}, nil
}
