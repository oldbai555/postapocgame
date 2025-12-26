// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUpdateLogic {
	return &FileUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileUpdateLogic) FileUpdate(req *types.FileUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "文件ID不能为空")
	}

	fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	file, err := fileRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询文件失败", err)
	}

	if req.Name != "" {
		file.Name = req.Name
	}
	// Status 字段：0 是有效值（禁用），需要特殊处理
	if req.Status == 0 || req.Status == 1 {
		file.Status = req.Status
	}

	if err := fileRepo.Update(l.ctx, file); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新文件记录失败", err)
	}
	return nil
}
