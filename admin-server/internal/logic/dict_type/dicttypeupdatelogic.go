// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict_type

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictTypeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeUpdateLogic {
	return &DictTypeUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictTypeUpdateLogic) DictTypeUpdate(req *types.DictTypeUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "字典类型ID不能为空")
	}

	dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
	dictType, err := dictTypeRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询字典类型失败", err)
	}

	if req.Name != "" {
		dictType.Name = req.Name
	}
	if req.Description != "" {
		dictType.Description = sql.NullString{String: req.Description, Valid: true}
	}
	// Status 字段：0 是有效值（禁用），需要特殊处理
	if req.Status == 0 || req.Status == 1 {
		dictType.Status = req.Status
	}

	if err := dictTypeRepo.Update(l.ctx, dictType); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新字典类型失败", err)
	}
	return nil
}
