// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package api

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApiCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApiCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiCreateLogic {
	return &ApiCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApiCreateLogic) ApiCreate(req *types.ApiCreateReq) error {
	if req == nil || req.Name == "" || req.Method == "" || req.Path == "" {
		return errs.New(errs.CodeBadRequest, "接口名称、方法和路径不能为空")
	}

	apiRepo := repository.NewApiRepository(l.svcCtx.Repository)
	// 检查是否已存在相同的 method+path
	_, err := apiRepo.FindByMethodAndPath(l.ctx, req.Method, req.Path)
	if err == nil {
		return errs.New(errs.CodeBadRequest, "该接口已存在")
	}

	api := model.AdminApi{
		Name:        req.Name,
		Method:      req.Method,
		Path:        req.Path,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:      req.Status,
	}
	if api.Status == 0 {
		api.Status = 1
	}

	if err := apiRepo.Create(l.ctx, &api); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建接口失败", err)
	}
	return nil
}
