// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"

	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"

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
	// todo: add your logic here and delete this line

	return nil
}
