// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"
	"sort"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuTreeLogic {
	return &MenuTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuTreeLogic) MenuTree() (resp *types.MenuTreeResp, err error) {
	// 尝试从缓存获取
	cache := l.svcCtx.Repository.BusinessCache
	var cachedResp types.MenuTreeResp
	err = cache.GetMenuTree(l.ctx, &cachedResp)
	if err == nil {
		return &cachedResp, nil
	}

	// 缓存未命中，从数据库查询
	menuRepo := repository.NewMenuRepository(l.svcCtx.Repository)
	list, err := menuRepo.ListAll(l.ctx)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询菜单列表失败", err)
	}

	// 构建 id → MenuItem 映射
	nodeMap := make(map[uint64]*types.MenuItem, len(list))
	var rootPtrs []*types.MenuItem

	// 第一遍：创建所有节点
	for _, m := range list {
		item := &types.MenuItem{
			Id:        m.Id,
			ParentId:  m.ParentId,
			Name:      m.Name,
			Path:      m.Path,
			Component: m.Component,
			Icon:      m.Icon,
			MenuType:  int64(m.Type),
			OrderNum:  m.OrderNum,
			Visible:   m.Visible,
			Status:    m.Status,
			Children:  []types.MenuItem{},
		}
		nodeMap[m.Id] = item
	}

	// 第二遍：构建树结构
	for _, item := range nodeMap {
		if item.ParentId == 0 {
			rootPtrs = append(rootPtrs, item)
			continue
		}
		if parent, ok := nodeMap[item.ParentId]; ok {
			parent.Children = append(parent.Children, *item)
		} else {
			// 父节点不存在，作为根节点处理
			rootPtrs = append(rootPtrs, item)
		}
	}

	// 转换为值类型并排序
	roots := make([]types.MenuItem, 0, len(rootPtrs))
	for _, ptr := range rootPtrs {
		roots = append(roots, *ptr)
	}

	// 对根节点和子节点按 orderNum 排序
	sortMenuItems(&roots)
	for i := range roots {
		sortMenuItems(&roots[i].Children)
	}

	resp = &types.MenuTreeResp{
		List: roots,
	}

	// 写入缓存（异步，不阻塞返回）
	go func() {
		if err := cache.SetMenuTree(context.Background(), resp); err != nil {
			l.Errorf("设置菜单树缓存失败: %v", err)
		}
	}()

	return resp, nil
}

// sortMenuItems 按 orderNum 和 id 排序菜单项
func sortMenuItems(items *[]types.MenuItem) {
	if items == nil || len(*items) == 0 {
		return
	}
	// 使用 sort.Slice 排序
	sort.Slice(*items, func(i, j int) bool {
		if (*items)[i].OrderNum != (*items)[j].OrderNum {
			return (*items)[i].OrderNum < (*items)[j].OrderNum
		}
		return (*items)[i].Id < (*items)[j].Id
	})
}
