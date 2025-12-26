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

type DictTypeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeDeleteLogic {
	return &DictTypeDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictTypeDeleteLogic) DictTypeDelete(req *types.DictTypeDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "字典类型ID不能为空")
	}

	// 检查是否有字典项关联
	dictItemRepo := repository.NewDictItemRepository(l.svcCtx.Repository)
	items, err := dictItemRepo.FindByTypeID(l.ctx, req.Id)
	if err == nil && len(items) > 0 {
		return errs.New(errs.CodeBadRequest, "该字典类型下存在字典项，无法删除")
	}

	dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
	if err := dictTypeRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除字典类型失败", err)
	}
	return nil
}
