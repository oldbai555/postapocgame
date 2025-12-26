// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileListLogic {
	return &FileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileListLogic) FileList(req *types.FileListReq) (resp *types.FileListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	list, total, err := fileRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Name)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询文件列表失败", err)
	}

	items := make([]types.FileItem, 0, len(list))
	for _, f := range list {
		items = append(items, types.FileItem{
			Id:        f.Id,
			Name:      f.Name,
			Status:    f.Status,
			CreatedAt: time.Unix(f.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		})
	}

	return &types.FileListResp{
		Total: total,
		List:  items,
	}, nil
}
