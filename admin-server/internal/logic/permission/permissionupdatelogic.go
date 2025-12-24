// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermissionUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PermissionUpdateLogic {
	return &PermissionUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermissionUpdateLogic) PermissionUpdate(req *types.PermissionUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "权限ID不能为空")
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	p, err := permissionRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询权限失败", err)
	}

	p.Name = req.Name
	p.Description = sql.NullString{String: req.Description, Valid: req.Description != ""}

	if err := permissionRepo.Update(l.ctx, p); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新权限失败", err)
	}
	return nil
}
