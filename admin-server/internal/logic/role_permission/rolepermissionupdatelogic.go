// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role_permission

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type RolePermissionUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRolePermissionUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePermissionUpdateLogic {
	return &RolePermissionUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RolePermissionUpdateLogic) RolePermissionUpdate(req *types.RolePermissionUpdateReq) error {
	if req.RoleId == 0 {
		return errs.New(errs.CodeBadRequest, "角色ID不能为空")
	}

	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	// 验证角色是否存在
	_, err := roleRepo.FindByID(l.ctx, req.RoleId)
	if err != nil {
		return errs.Wrap(errs.CodeBadRequest, "角色不存在", err)
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	// 验证所有权限是否存在
	for _, permID := range req.PermissionIds {
		_, err := permissionRepo.FindByID(l.ctx, permID)
		if err != nil {
			return errs.Wrap(errs.CodeBadRequest, "权限不存在", err)
		}
	}

	rolePermissionRepo := repository.NewRolePermissionRepository(l.svcCtx.Repository)
	if err := rolePermissionRepo.UpdateRolePermissions(l.ctx, req.RoleId, req.PermissionIds); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新角色权限失败", err)
	}

	// 清除所有拥有该角色的用户的权限和菜单树缓存
	// 注意：由于 go-zero Redis 不支持 SCAN，这里只能清除已知的缓存
	// 实际场景中，可以通过定时任务或延迟清除策略来处理
	// 这里先清除完整菜单树缓存，用户权限缓存会在下次查询时自动更新
	cache := l.svcCtx.Repository.BusinessCache
	go func() {
		// 清除完整菜单树缓存（因为权限变更可能影响菜单显示）
		if err := cache.DeleteMenuTree(context.Background()); err != nil {
			l.Errorf("清除菜单树缓存失败: %v", err)
		}
		// 注意：无法直接清除所有用户的权限缓存，需要在查询时检查并更新
	}()

	return nil
}
