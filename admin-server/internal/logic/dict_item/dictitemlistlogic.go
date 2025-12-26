// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict_item

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictItemListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictItemListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictItemListLogic {
	return &DictItemListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictItemListLogic) DictItemList(req *types.DictItemListReq) (resp *types.DictItemListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	dictItemRepo := repository.NewDictItemRepository(l.svcCtx.Repository)
	list, total, err := dictItemRepo.FindPage(l.ctx, req.Page, req.PageSize, req.TypeId, req.Label)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询字典项列表失败", err)
	}

	items := make([]types.DictItemItem, 0, len(list))
	for _, di := range list {
		remark := ""
		if di.Remark.Valid {
			remark = di.Remark.String
		}
		items = append(items, types.DictItemItem{
			Id:        di.Id,
			TypeId:    di.TypeId,
			Label:     di.Label,
			Value:     di.Value,
			Sort:      di.Sort,
			Status:    di.Status,
			Remark:    remark,
			CreatedAt: time.Unix(di.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		})
	}

	return &types.DictItemListResp{
		Total: total,
		List:  items,
	}, nil
}
