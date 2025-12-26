// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
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

func (l *FileDownloadLogic) FileDownload(w http.ResponseWriter, r *http.Request, idStr string) error {
	if idStr == "" {
		return errs.New(errs.CodeBadRequest, "文件ID不能为空")
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return errs.Wrap(errs.CodeBadRequest, "文件ID格式错误", err)
	}

	fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	file, err := fileRepo.FindByID(l.ctx, id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询文件失败", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(file.Path); os.IsNotExist(err) {
		return errs.New(errs.CodeNotFound, "文件不存在")
	}

	// 设置响应头，直接下载文件
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalName))
	if file.MimeType.Valid && file.MimeType.String != "" {
		w.Header().Set("Content-Type", file.MimeType.String)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Length", strconv.FormatUint(file.Size, 10))

	// 直接返回文件内容
	http.ServeFile(w, r, file.Path)
	return nil
}
