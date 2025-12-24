// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission_api

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionApiUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermissionApiUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PermissionApiUpdateLogic {
	return &PermissionApiUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermissionApiUpdateLogic) PermissionApiUpdate(req *types.PermissionApiUpdateReq) error {
	if req.PermissionId == 0 {
		return errs.New(errs.CodeBadRequest, "权限ID不能为空")
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	// 验证权限是否存在
	_, err := permissionRepo.FindByID(l.ctx, req.PermissionId)
	if err != nil {
		return errs.Wrap(errs.CodeBadRequest, "权限不存在", err)
	}

	apiRepo := repository.NewApiRepository(l.svcCtx.Repository)
	// 验证所有接口是否存在
	for _, apiID := range req.ApiIds {
		_, err := apiRepo.FindByID(l.ctx, apiID)
		if err != nil {
			return errs.Wrap(errs.CodeBadRequest, "接口不存在", err)
		}
	}

	permissionApiRepo := repository.NewPermissionApiRepository(l.svcCtx.Repository)
	if err := permissionApiRepo.UpdatePermissionApis(l.ctx, req.PermissionId, req.ApiIds); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新权限接口失败", err)
	}
	return nil
}
