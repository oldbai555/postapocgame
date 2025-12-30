// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict_type

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictTypeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeListLogic {
	return &DictTypeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictTypeListLogic) DictTypeList(req *types.DictTypeListReq) (resp *types.DictTypeListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
	list, total, err := dictTypeRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Name, req.Code)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询字典类型列表失败", err)
	}

	items := make([]types.DictTypeItem, 0, len(list))
	for _, dt := range list {
		description := ""
		if dt.Description.Valid {
			description = dt.Description.String
		}
		items = append(items, types.DictTypeItem{
			Id:          dt.Id,
			Name:        dt.Name,
			Code:        dt.Code,
			Description: description,
			Status:      dt.Status,
			CreatedAt:   dt.CreatedAt,
		})
	}

	return &types.DictTypeListResp{
		Total: total,
		List:  items,
	}, nil
}
