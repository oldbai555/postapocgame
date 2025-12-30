// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notice

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type NoticeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeUpdateLogic {
	return &NoticeUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NoticeUpdateLogic) NoticeUpdate(req *types.NoticeUpdateReq) (resp *types.Response, err error) {
	if req == nil || req.Id == 0 {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	noticeRepo := repository.NewNoticeRepository(l.svcCtx.Repository)
	notice, err := noticeRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeNotFound, "公告不存在", err)
	}

	// 更新字段
	if req.Title != "" {
		notice.Title = req.Title
	}
	if req.Content != "" {
		notice.Content = req.Content
	}
	if req.NoticeType > 0 {
		notice.Type = req.NoticeType
	}
	// 状态：1=草稿，2=已发布，只有大于0才更新
	if req.Status > 0 {
		notice.Status = req.Status
	}
	if req.PublishTime > 0 {
		notice.PublishTime = req.PublishTime
	}

	// 保存原始状态（在更新之前）
	oldStatus := notice.Status

	notice.UpdatedAt = time.Now().Unix()

	if err := noticeRepo.Update(l.ctx, notice); err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "更新公告失败", err)
	}

	// 如果状态从草稿(1)变为已发布(2)，给所有用户创建通知
	if oldStatus == 1 && notice.Status == 2 {
		go l.createNotificationsForAllUsers(notice.Id, notice.Title, notice.Content)
	}

	return &types.Response{
		Code:    0,
		Message: "更新成功",
	}, nil
}

// createNotificationsForAllUsers 给所有用户创建公告通知
func (l *NoticeUpdateLogic) createNotificationsForAllUsers(noticeID uint64, title, content string) {
	defer func() {
		if r := recover(); r != nil {
			l.Errorf("创建公告通知时发生 panic: %v, noticeId=%d", r, noticeID)
		}
	}()

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	notificationRepo := repository.NewNotificationRepository(l.svcCtx.Repository)

	// 分批获取所有用户
	limit := int64(100)
	lastID := uint64(0)
	totalCreated := 0

	for {
		users, newLastID, err := userRepo.FindChunk(context.Background(), limit, lastID)
		if err != nil {
			l.Errorf("查询用户失败: noticeId=%d, error: %v", noticeID, err)
			break
		}

		if len(users) == 0 {
			break
		}

		now := time.Now().Unix()
		for _, user := range users {
			// 检查是否已存在通知（避免重复创建）
			notifications, _, err := notificationRepo.FindPage(context.Background(), 1, 1, user.Id, "notice", -1)
			if err == nil {
				hasNotification := false
				for _, notif := range notifications {
					if notif.SourceId == noticeID && notif.SourceType == "notice" && notif.DeletedAt == 0 {
						hasNotification = true
						break
					}
				}
				if hasNotification {
					continue
				}
			}

			// 创建通知
			notification := &model.AdminNotification{
				UserId:     user.Id,
				SourceType: "notice",
				SourceId:   noticeID,
				Title:      title,
				Content:    content,
				ReadStatus: 0, // 未读
				ReadAt:     0,
				CreatedAt:  now,
				UpdatedAt:  now,
				DeletedAt:  0,
			}

			if err := notificationRepo.Create(context.Background(), notification); err != nil {
				l.Errorf("创建公告通知失败: userId=%d, noticeId=%d, error: %v", user.Id, noticeID, err)
			} else {
				totalCreated++
			}
		}

		if len(users) < int(limit) {
			break
		}
		lastID = newLastID
	}

	l.Infof("成功为公告创建通知: noticeId=%d, totalCreated=%d", noticeID, totalCreated)
}
