// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notice

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/consts"
	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type NoticeCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeCreateLogic {
	return &NoticeCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NoticeCreateLogic) NoticeCreate(req *types.NoticeCreateReq) (resp *types.Response, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}
	if req.Title == "" {
		return nil, errs.New(errs.CodeBadRequest, "公告标题不能为空")
	}
	if req.Content == "" {
		return nil, errs.New(errs.CodeBadRequest, "公告内容不能为空")
	}

	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 设置默认值
	noticeType := req.NoticeType
	if noticeType == 0 {
		noticeType = 1 // 默认普通公告
	}
	status := req.Status
	if status == 0 {
		status = consts.NoticeStatusDraft // 默认草稿
	}
	publishTime := req.PublishTime
	if publishTime == 0 {
		publishTime = time.Now().Unix() // 默认立即发布
	}

	notice := &model.AdminNotice{
		Title:       req.Title,
		Content:     req.Content,
		Type:        noticeType,
		Status:      status,
		PublishTime: publishTime,
		CreatedBy:   user.UserID,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		DeletedAt:   0,
	}

	noticeRepo := repository.NewNoticeRepository(l.svcCtx.Repository)
	if err := noticeRepo.Create(l.ctx, notice); err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "创建公告失败", err)
	}

	// 如果状态是已发布，给所有用户创建通知
	if status == consts.NoticeStatusPublished {
		go l.createNotificationsForAllUsers(notice.Id, notice.Title, notice.Content)
	}

	return &types.Response{
		Code:    0,
		Message: "创建成功",
	}, nil
}

// createNotificationsForAllUsers 给所有用户创建公告通知
func (l *NoticeCreateLogic) createNotificationsForAllUsers(noticeID uint64, title, content string) {
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
