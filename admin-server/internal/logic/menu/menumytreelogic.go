// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuMyTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuMyTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuMyTreeLogic {
	return &MenuMyTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuMyTreeLogic) MenuMyTree() (resp *types.MenuTreeResp, err error) {
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 超级管理员（user_id=1）默认拥有最高权限，直接返回完整菜单树
	if user.UserID == 1 {
		treeLogic := NewMenuTreeLogic(l.ctx, l.svcCtx)
		return treeLogic.MenuTree()
	}

	// 尝试从缓存获取用户菜单树
	cache := l.svcCtx.Repository.BusinessCache
	var cachedResp types.MenuTreeResp
	err = cache.GetUserMenuTree(l.ctx, user.UserID, &cachedResp)
	if err == nil {
		return &cachedResp, nil
	}

	// 获取用户权限编码
	userRoleRepo := repository.NewUserRoleRepository(l.svcCtx.Repository)
	roleIDs, err := userRoleRepo.ListRoleIDsByUserID(l.ctx, user.UserID)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询用户角色失败", err)
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	perms, err := permissionRepo.ListByRoleIDs(l.ctx, roleIDs)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询用户权限失败", err)
	}

	permCodes := make([]string, 0, len(perms))
	permSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		permCodes = append(permCodes, p.Code)
		permSet[p.Code] = struct{}{}
	}

	// 检查是否是超级管理员（拥有 * 权限）
	hasSuperAdmin := false
	for _, code := range permCodes {
		if code == "*" {
			hasSuperAdmin = true
			break
		}
	}

	// 如果是超级管理员，直接返回完整菜单树
	if hasSuperAdmin {
		treeLogic := NewMenuTreeLogic(l.ctx, l.svcCtx)
		return treeLogic.MenuTree()
	}

	// 获取「菜单ID -> 绑定的权限编码列表」的完整映射
	permissionMenuRepo := repository.NewPermissionMenuRepository(l.svcCtx.Repository)
	menuPermissionMap, err := permissionMenuRepo.ListMenuPermissionCodes(l.ctx)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询菜单权限关联失败", err)
	}

	// 获取完整菜单树
	treeLogic := NewMenuTreeLogic(l.ctx, l.svcCtx)
	fullTree, err := treeLogic.MenuTree()
	if err != nil {
		return nil, err
	}

	// 递归过滤菜单树：只保留有权限的菜单/按钮，或没有绑定权限的菜单（目录通常不需要权限）
	var filterMenu func(item types.MenuItem) *types.MenuItem
	filterMenu = func(item types.MenuItem) *types.MenuItem {
		// 如果菜单绑定了权限，检查用户是否有该权限
		if permCodes, hasPerms := menuPermissionMap[item.Id]; hasPerms && len(permCodes) > 0 {
			hasAccess := false
			for _, code := range permCodes {
				if _, ok := permSet[code]; ok {
					hasAccess = true
					break
				}
			}
			if !hasAccess {
				return nil // 无权限，过滤掉
			}
		}
		// 如果菜单没有绑定权限，或者用户有权限，继续处理子菜单
		filtered := item
		filtered.Children = []types.MenuItem{}
		for _, child := range item.Children {
			if filteredChild := filterMenu(child); filteredChild != nil {
				filtered.Children = append(filtered.Children, *filteredChild)
			}
		}
		// 如果是按钮，必须有权限才保留
		if item.MenuType == 3 {
			if permCodes, hasPerms := menuPermissionMap[item.Id]; hasPerms && len(permCodes) > 0 {
				hasAccess := false
				for _, code := range permCodes {
					if _, ok := permSet[code]; ok {
						hasAccess = true
						break
					}
				}
				if !hasAccess {
					return nil
				}
			}
		}
		return &filtered
	}

	filteredRoots := make([]types.MenuItem, 0)
	for _, root := range fullTree.List {
		if filtered := filterMenu(root); filtered != nil {
			filteredRoots = append(filteredRoots, *filtered)
		}
	}

	resp = &types.MenuTreeResp{
		List: filteredRoots,
	}

	// 写入缓存（异步，不阻塞返回）
	go func() {
		if err := cache.SetUserMenuTree(context.Background(), user.UserID, resp); err != nil {
			l.Errorf("设置用户菜单树缓存失败: userId=%d, error=%v", user.UserID, err)
		}
	}()

	return resp, nil
}
