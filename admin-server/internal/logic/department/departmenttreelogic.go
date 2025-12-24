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

type DepartmentTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDepartmentTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepartmentTreeLogic {
	return &DepartmentTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DepartmentTreeLogic) DepartmentTree() (resp *types.DepartmentTreeResp, err error) {
	deptRepo := repository.NewDepartmentRepository(l.svcCtx.Repository)
	list, err := deptRepo.ListAll(l.ctx)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询部门列表失败", err)
	}

	// 构建 id → DepartmentItem 映射
	nodeMap := make(map[uint64]*types.DepartmentItem, len(list))
	var roots []types.DepartmentItem

	for _, d := range list {
		item := types.DepartmentItem{
			Id:       d.Id,
			ParentId: d.ParentId,
			Name:     d.Name,
			OrderNum: d.OrderNum,
			Status:   d.Status,
			Children: []types.DepartmentItem{},
		}
		nodeMap[d.Id] = &item
	}

	for _, item := range nodeMap {
		if item.ParentId == 0 {
			roots = append(roots, *item)
			continue
		}
		if parent, ok := nodeMap[item.ParentId]; ok {
			parent.Children = append(parent.Children, *item)
		} else {
			roots = append(roots, *item)
		}
	}

	return &types.DepartmentTreeResp{
		List: roots,
	}, nil
}
