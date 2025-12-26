// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileCreateLogic {
	return &FileCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileCreateLogic) FileCreate(req *types.FileCreateReq) error {
	if req == nil || req.Name == "" {
		return errs.New(errs.CodeBadRequest, "文件名称不能为空")
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	file := model.AdminFile{
		Name:         req.Name,
		OriginalName: req.Name, // 默认使用 name 作为原始名称
		Path:         "",       // 文件路径需要在上传时设置
		Url:          "",       // 文件URL需要在上传时设置
		Size:         0,        // 文件大小需要在上传时设置
		MimeType:     sql.NullString{Valid: false},
		Ext:          sql.NullString{Valid: false},
		StorageType:  "local", // 默认本地存储
		Status:       status,
	}

	fileRepo := repository.NewFileRepository(l.svcCtx.Repository)
	if err := fileRepo.Create(l.ctx, &file); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建文件记录失败", err)
	}
	return nil
}
