// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict_type

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

type DictTypeCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeCreateLogic {
	return &DictTypeCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictTypeCreateLogic) DictTypeCreate(req *types.DictTypeCreateReq) error {
	if req == nil || req.Name == "" || req.Code == "" {
		return errs.New(errs.CodeBadRequest, "字典类型名称和编码不能为空")
	}

	dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
	// 检查编码是否已存在
	_, err := dictTypeRepo.FindByCode(l.ctx, req.Code)
	if err == nil {
		return errs.New(errs.CodeBadRequest, "字典类型编码已存在")
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	dictType := model.AdminDictType{
		Name:        req.Name,
		Code:        req.Code,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:      status,
	}

	if err := dictTypeRepo.Create(l.ctx, &dictType); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建字典类型失败", err)
	}
	return nil
}
