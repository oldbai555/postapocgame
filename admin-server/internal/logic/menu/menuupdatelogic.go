// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuUpdateLogic {
	return &MenuUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuUpdateLogic) MenuUpdate(req *types.MenuUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "菜单ID不能为空")
	}

	menuRepo := repository.NewMenuRepository(l.svcCtx.Repository)
	m, err := menuRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询菜单失败", err)
	}

	m.ParentId = req.ParentId
	m.Name = req.Name
	m.Path = req.Path
	m.Component = req.Component
	m.Icon = req.Icon
	m.Type = req.MenuType
	m.OrderNum = req.OrderNum
	if req.Visible != 0 {
		m.Visible = req.Visible
	}
	if req.Status != 0 {
		m.Status = req.Status
	}

	if err := menuRepo.Update(l.ctx, m); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新菜单失败", err)
	}
	return nil
}
