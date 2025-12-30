// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package demo

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DemoListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDemoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DemoListLogic {
	return &DemoListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DemoListLogic) DemoList(req *types.DemoListReq) (resp *types.DemoListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	demoRepo := repository.NewDemoRepository(l.svcCtx.Repository)
	list, total, err := demoRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Name)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询演示功能列表失败", err)
	}

	items := make([]types.DemoItem, 0, len(list))
	for _, d := range list {
		items = append(items, types.DemoItem{
			Id:        d.Id,
			Name:      d.Name,
			Status:    d.Status,
			CreatedAt: d.CreatedAt,
		})
	}

	return &types.DemoListResp{
		Total: total,
		List:  items,
	}, nil
}
