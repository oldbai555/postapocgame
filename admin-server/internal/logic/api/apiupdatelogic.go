// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package api

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApiUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApiUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiUpdateLogic {
	return &ApiUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApiUpdateLogic) ApiUpdate(req *types.ApiUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "接口ID不能为空")
	}

	apiRepo := repository.NewApiRepository(l.svcCtx.Repository)
	api, err := apiRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询接口失败", err)
	}

	if req.Name != "" {
		api.Name = req.Name
	}
	if req.Method != "" {
		api.Method = req.Method
	}
	if req.Path != "" {
		api.Path = req.Path
	}
	if req.Description != "" {
		api.Description = sql.NullString{String: req.Description, Valid: true}
	}
	// Status 字段：0 是有效值（禁用），需要特殊处理
	if req.Status == 0 || req.Status == 1 {
		api.Status = req.Status
	}

	if err := apiRepo.Update(l.ctx, api); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新接口失败", err)
	}
	return nil
}
