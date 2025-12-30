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

type NotificationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationListLogic {
	return &NotificationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotificationListLogic) NotificationList(req *types.NotificationListReq) (resp *types.NotificationListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 只查询当前用户的消息通知
	notificationRepo := repository.NewNotificationRepository(l.svcCtx.Repository)
	list, total, err := notificationRepo.FindPage(l.ctx, req.Page, req.PageSize, user.UserID, req.SourceType, req.ReadStatus)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询消息通知列表失败", err)
	}

	items := make([]types.NotificationItem, 0, len(list))
	for _, n := range list {
		items = append(items, types.NotificationItem{
			Id:         n.Id,
			UserId:     n.UserId,
			SourceType: n.SourceType,
			SourceId:   n.SourceId,
			Title:      n.Title,
			Content:    n.Content,
			ReadStatus: n.ReadStatus,
			ReadAt:     n.ReadAt,
			CreatedAt:  n.CreatedAt,
			UpdatedAt:  n.UpdatedAt,
		})
	}

	return &types.NotificationListResp{
		Total: total,
		List:  items,
	}, nil
}
