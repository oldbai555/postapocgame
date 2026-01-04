// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"fmt"
	"os"
	"strings"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileDownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileDownloadLogic {
	return &FileDownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileDownloadLogic) FileDownload(req *types.FileDownloadReq) (resp *types.FileDownloadResp, err error) {
	if req == nil || req.Id == 0 {
		return nil, errs.New(errs.CodeBadRequest, "文件ID不能为空")
	}

	fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	file, err := fileRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeNotFound, "文件不存在", err)
	}

	// 检查文件是否存在
	fileSystemPath := file.Path
	if strings.HasPrefix(file.Path, "/uploads/") {
		fileSystemPath = "." + file.Path
	} else if !strings.HasPrefix(file.Path, "./") {
		fileSystemPath = "./" + file.Path
	}

	if _, err := os.Stat(fileSystemPath); os.IsNotExist(err) {
		return nil, errs.New(errs.CodeNotFound, "文件不存在")
	}

	// 构建文件访问URL（使用 /api/v1/uploads/xxx 格式，前端可以通过代理访问）
	// file.Path 格式：/uploads/xxx 或 /api/v1/uploads/xxx
	accessPath := file.Path
	if strings.HasPrefix(file.Path, "/uploads/") {
		// 如果是 /uploads/xxx，转换为 /api/v1/uploads/xxx
		accessPath = fmt.Sprintf("/api/v1%s", file.Path)
	} else if !strings.HasPrefix(file.Path, "/api/v1/") {
		// 如果既不是 /uploads/ 也不是 /api/v1/，添加 /api/v1/uploads/ 前缀
		accessPath = fmt.Sprintf("/api/v1/uploads/%s", strings.TrimPrefix(file.Path, "/"))
	}

	return &types.FileDownloadResp{
		Url: accessPath,
	}, nil
}
