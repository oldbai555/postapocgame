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

	fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	// 先查询文件信息，获取文件路径（后续可以扩展删除物理文件）
	_, err := fileRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询文件失败", err)
	}

	// TODO: 后续可以扩展删除物理文件的功能
	// if file.StorageType == "local" && file.Path != "" {
	//     // 删除本地文件
	//     os.Remove(file.Path)
	// }

	// 删除数据库记录（软删除）
	if err := fileRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除文件记录失败", err)
	}
	return nil
}
