// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"

	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileDeleteLogic {
	return &FileDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileDeleteLogic) FileDelete(req *types.FileDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "文件ID不能为空")
	}

	// TODO: 实现 FileRepository 后，使用以下代码
	// fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	// if err := fileRepo.DeleteByID(l.ctx, req.Id); err != nil {
	// 	return errs.Wrap(errs.CodeInternalError, "删除文件失败", err)
	// }
	// return nil

	return errs.New(errs.CodeInternalError, "文件删除功能待实现")
}
