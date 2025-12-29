// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	"postapocgame/admin-server/pkg/initdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuDeleteLogic {
	return &MenuDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuDeleteLogic) MenuDelete(req *types.MenuDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "菜单ID不能为空")
	}
	// 保护初始化数据：不允许删除初始化菜单
	if initdata.IsInitMenuID(req.Id) {
		return errs.New(errs.CodeBadRequest, "初始化数据不可删除")
	}

	menuRepo := repository.NewMenuRepository(l.svcCtx.Repository)
	if err := menuRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除菜单失败", err)
	}

	// 清除菜单树缓存
	cache := l.svcCtx.Repository.BusinessCache
	go func() {
		if err := cache.DeleteMenuTree(context.Background()); err != nil {
			l.Errorf("清除菜单树缓存失败: %v", err)
		}
	}()

	return nil
}
