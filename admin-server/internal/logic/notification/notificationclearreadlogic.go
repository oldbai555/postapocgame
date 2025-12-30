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

type NotificationClearReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationClearReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationClearReadLogic {
	return &NotificationClearReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotificationClearReadLogic) NotificationClearRead() (resp *types.Response, err error) {
	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	notificationRepo := repository.NewNotificationRepository(l.svcCtx.Repository)
	if err := notificationRepo.ClearRead(l.ctx, user.UserID); err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "清除已读消息失败", err)
	}

	return &types.Response{
		Code:    0,
		Message: "操作成功",
	}, nil
}
