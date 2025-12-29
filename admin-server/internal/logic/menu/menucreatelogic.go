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

	// 清除菜单树缓存
	cache := l.svcCtx.Repository.BusinessCache
	go func() {
		if err := cache.DeleteMenuTree(context.Background()); err != nil {
			l.Errorf("清除菜单树缓存失败: %v", err)
		}
		// 清除所有用户的菜单树缓存（因为菜单变更会影响所有用户）
		// 注意：go-zero Redis 不支持 SCAN，这里只能清除已知的缓存
		// 实际场景中，可以通过定时任务或延迟清除策略来处理
	}()

	return nil
}
