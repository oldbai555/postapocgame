// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DepartmentUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDepartmentUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepartmentUpdateLogic {
	return &DepartmentUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DepartmentUpdateLogic) DepartmentUpdate(req *types.DepartmentUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "部门ID不能为空")
	}

	deptRepo := repository.NewDepartmentRepository(l.svcCtx.Repository)
	dept, err := deptRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询部门失败", err)
	}

	dept.ParentId = req.ParentId
	dept.Name = req.Name
	if req.OrderNum != 0 {
		dept.OrderNum = req.OrderNum
	}
	if req.Status != 0 {
		dept.Status = req.Status
	}

	if err := deptRepo.Update(l.ctx, dept); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新部门失败", err)
	}
	return nil
}
