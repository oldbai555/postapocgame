// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	"postapocgame/admin-server/pkg/initdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type DepartmentDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDepartmentDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepartmentDeleteLogic {
	return &DepartmentDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DepartmentDeleteLogic) DepartmentDelete(req *types.DepartmentDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "部门ID不能为空")
	}
	// 保护初始化数据：不允许删除根部门（id=1）
	if initdata.IsInitDepartmentID(req.Id) {
		return errs.New(errs.CodeBadRequest, "初始化数据不可删除")
	}

	deptRepo := repository.NewDepartmentRepository(l.svcCtx.Repository)
	if err := deptRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除部门失败", err)
	}
	return nil
}
