// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"context"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DepartmentCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDepartmentCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepartmentCreateLogic {
	return &DepartmentCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DepartmentCreateLogic) DepartmentCreate(req *types.DepartmentCreateReq) error {
	if req == nil || req.Name == "" {
		return errs.New(errs.CodeBadRequest, "部门名称不能为空")
	}

	deptRepo := repository.NewDepartmentRepository(l.svcCtx.Repository)
	dept := model.AdminDepartment{
		ParentId: req.ParentId,
		Name:     req.Name,
		OrderNum: req.OrderNum,
		Status:   req.Status,
	}

	if err := deptRepo.Create(l.ctx, &dept); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建部门失败", err)
	}
	return nil
}
