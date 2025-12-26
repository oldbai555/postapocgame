// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict_item

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictItemDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictItemDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictItemDeleteLogic {
	return &DictItemDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictItemDeleteLogic) DictItemDelete(req *types.DictItemDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "字典项ID不能为空")
	}

	dictItemRepo := repository.NewDictItemRepository(l.svcCtx.Repository)
	if err := dictItemRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除字典项失败", err)
	}
	return nil
}
