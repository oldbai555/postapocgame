// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadLogic {
	return &FileUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileUploadLogic) FileUpload(r *http.Request) (resp *types.FileUploadResp, err error) {
	// 解析 multipart/form-data
	err = r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		return nil, errs.Wrap(errs.CodeBadRequest, "解析上传文件失败", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, errs.Wrap(errs.CodeBadRequest, "获取上传文件失败", err)
	}
	defer file.Close()

	// 创建上传目录（如果不存在）
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "创建上传目录失败", err)
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	filePath := filepath.Join(uploadDir, fileName)
	fileURL := fmt.Sprintf("/uploads/%s", fileName)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "创建文件失败", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "保存文件失败", err)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "获取文件信息失败", err)
	}

	// 获取 MIME 类型
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType([]byte(ext))
	}

	// 保存文件记录到数据库
	fileModel := model.AdminFile{
		Name:         fileName,
		OriginalName: header.Filename,
		Path:         filePath,
		Url:          fileURL,
		Size:         uint64(fileInfo.Size()),
		MimeType:     sql.NullString{String: mimeType, Valid: mimeType != ""},
		Ext:          sql.NullString{String: strings.TrimPrefix(ext, "."), Valid: ext != ""},
		StorageType:  "local",
		Status:       1,
	}

	fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	if err := fileRepo.Create(l.ctx, &fileModel); err != nil {
		// 如果数据库保存失败，删除已上传的文件
		os.Remove(filePath)
		return nil, errs.Wrap(errs.CodeInternalError, "保存文件记录失败", err)
	}

	return &types.FileUploadResp{
		Id:           fileModel.Id,
		Name:         fileModel.Name,
		OriginalName: fileModel.OriginalName,
		Path:         fileModel.Path,
		Url:          fileModel.Url,
		Size:         fileModel.Size,
		MimeType:     mimeType,
		Ext:          strings.TrimPrefix(ext, "."),
	}, nil
}
