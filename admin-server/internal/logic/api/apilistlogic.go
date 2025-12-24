// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package api

import (
	"context"
	"strconv"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApiListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApiListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiListLogic {
	return &ApiListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApiListLogic) ApiList(req *types.ApiListReq) (resp *types.ApiListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	apiRepo := repository.NewApiRepository(l.svcCtx.Repository)
	list, total, err := apiRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Name)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询接口列表失败", err)
	}

	items := make([]types.ApiItem, 0, len(list))
	for _, a := range list {
		createdAtStr := ""
		if a.CreatedAt > 0 {
			createdAtStr = strconv.FormatInt(int64(a.CreatedAt), 10)
		}
		description := ""
		if a.Description.Valid {
			description = a.Description.String
		}
		items = append(items, types.ApiItem{
			Id:          a.Id,
			Name:        a.Name,
			Method:      a.Method,
			Path:        a.Path,
			Description: description,
			Status:      a.Status,
			CreatedAt:   createdAtStr,
		})
	}

	return &types.ApiListResp{
		Total: total,
		List:  items,
	}, nil
}
