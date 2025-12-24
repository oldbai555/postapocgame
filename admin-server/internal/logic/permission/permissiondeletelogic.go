// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	"postapocgame/admin-server/pkg/initdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermissionDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PermissionDeleteLogic {
	return &PermissionDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermissionDeleteLogic) PermissionDelete(req *types.PermissionDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "权限ID不能为空")
	}
	// 保护初始化数据：不允许删除初始化权限
	if initdata.IsInitPermissionID(req.Id) {
		return errs.New(errs.CodeBadRequest, "初始化数据不可删除")
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	if err := permissionRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除权限失败", err)
	}
	return nil
}
