// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictGetLogic {
	return &DictGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictGetLogic) DictGet(req *types.DictGetReq) (resp *types.DictGetResp, err error) {
	if req == nil || req.Code == "" {
		return nil, errs.New(errs.CodeBadRequest, "字典类型编码不能为空")
	}

	dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
	dictType, err := dictTypeRepo.FindByCode(l.ctx, req.Code)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询字典类型失败", err)
	}

	dictItemRepo := repository.NewDictItemRepository(l.svcCtx.Repository)
	items, err := dictItemRepo.FindByTypeID(l.ctx, dictType.Id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询字典项失败", err)
	}

	dictItems := make([]types.DictItemItem, 0, len(items))
	for _, di := range items {
		remark := ""
		if di.Remark.Valid {
			remark = di.Remark.String
		}
		dictItems = append(dictItems, types.DictItemItem{
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

	return &types.DictGetResp{
		Code:  dictType.Code,
		Items: dictItems,
	}, nil
}
