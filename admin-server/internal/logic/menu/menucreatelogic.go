// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuCreateLogic {
	return &MenuCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuCreateLogic) MenuCreate(req *types.MenuCreateReq) error {
	if req == nil || req.Name == "" || req.MenuType == 0 {
		return errs.New(errs.CodeBadRequest, "菜单名称和类型不能为空")
	}

	menuRepo := repository.NewMenuRepository(l.svcCtx.Repository)
	m := model.AdminMenu{
		ParentId:  req.ParentId,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		Type:      req.MenuType,
		OrderNum:  req.OrderNum,
		Visible:   req.Visible,
		Status:    req.Status,
	}

	if err := menuRepo.Create(l.ctx, &m); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建菜单失败", err)
	}
	return nil
}
