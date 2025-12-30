// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type NotificationDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationDeleteLogic {
	return &NotificationDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotificationDeleteLogic) NotificationDelete(req *types.NotificationDeleteReq) (resp *types.Response, err error) {
	if req == nil || req.Id == 0 {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 验证消息通知是否属于当前用户
	notificationRepo := repository.NewNotificationRepository(l.svcCtx.Repository)
	notification, err := notificationRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeNotFound, "消息通知不存在", err)
	}
	if notification.UserId != user.UserID {
		return nil, errs.New(errs.CodeForbidden, "无权删除该消息通知")
	}

	if err := notificationRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "删除消息通知失败", err)
	}

	return &types.Response{
		Code:    0,
		Message: "删除成功",
	}, nil
}
