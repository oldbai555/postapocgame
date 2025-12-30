// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type NotificationReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationReadLogic {
	return &NotificationReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotificationReadLogic) NotificationRead(req *types.NotificationReadReq) (resp *types.Response, err error) {
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
		return nil, errs.New(errs.CodeForbidden, "无权操作该消息通知")
	}

	// 标记为已读
	now := time.Now().Unix()
	notification.ReadStatus = 1
	notification.ReadAt = now
	notification.UpdatedAt = now

	if err := notificationRepo.Update(l.ctx, notification); err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "标记已读失败", err)
	}

	return &types.Response{
		Code:    0,
		Message: "操作成功",
	}, nil
}
