// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notice

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type NoticeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeDeleteLogic {
	return &NoticeDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NoticeDeleteLogic) NoticeDelete(req *types.NoticeDeleteReq) (resp *types.Response, err error) {
	if req == nil || req.Id == 0 {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	noticeRepo := repository.NewNoticeRepository(l.svcCtx.Repository)
	if err := noticeRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "删除公告失败", err)
	}

	return &types.Response{
		Code:    0,
		Message: "删除成功",
	}, nil
}
