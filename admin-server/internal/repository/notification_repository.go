package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type NotificationRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminNotification, error)
	FindPage(ctx context.Context, page, pageSize int64, userID uint64, sourceType string, readStatus int64) ([]model.AdminNotification, int64, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, notification *model.AdminNotification) error
	Update(ctx context.Context, notification *model.AdminNotification) error
	// 全部已读：更新用户的所有未读消息为已读
	MarkAllAsRead(ctx context.Context, userID uint64) error
	// 清除已读：删除用户的所有已读消息
	ClearRead(ctx context.Context, userID uint64) error
}

type notificationRepository struct {
	model model.AdminNotificationModel
	conn  sqlx.SqlConn
}

func NewNotificationRepository(repo *Repository) NotificationRepository {
	return &notificationRepository{model: repo.AdminNotificationModel, conn: repo.DB}
}

func (r *notificationRepository) FindByID(ctx context.Context, id uint64) (*model.AdminNotification, error) {
	return r.model.FindOne(ctx, id)
}

func (r *notificationRepository) FindPage(ctx context.Context, page, pageSize int64, userID uint64, sourceType string, readStatus int64) ([]model.AdminNotification, int64, error) {
	// 构建查询条件
	where := []string{"deleted_at = 0"}
	args := []interface{}{}

	if userID > 0 {
		where = append(where, "user_id = ?")
		args = append(args, userID)
	}
	if sourceType != "" {
		where = append(where, "source_type = ?")
		args = append(args, sourceType)
	}
	if readStatus >= 0 {
		where = append(where, "read_status = ?")
		args = append(args, readStatus)
	}

	whereClause := strings.Join(where, " AND ")

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `admin_notification` WHERE %s", whereClause)
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	var list []model.AdminNotification
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT id, user_id, source_type, source_id, title, content, read_status, read_at, created_at, updated_at, deleted_at FROM `admin_notification` WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, pageSize, offset)
	err = r.conn.QueryRowsCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *notificationRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *notificationRepository) Create(ctx context.Context, notification *model.AdminNotification) error {
	_, err := r.model.Insert(ctx, notification)
	return err
}

func (r *notificationRepository) Update(ctx context.Context, notification *model.AdminNotification) error {
	return r.model.Update(ctx, notification)
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uint64) error {
	now := time.Now().Unix()
	query := "UPDATE `admin_notification` SET `read_status` = 1, `read_at` = ?, `updated_at` = ? WHERE `user_id` = ? AND `read_status` = 0 AND `deleted_at` = 0"
	_, err := r.conn.ExecCtx(ctx, query, now, now, userID)
	return err
}

func (r *notificationRepository) ClearRead(ctx context.Context, userID uint64) error {
	// 软删除所有已读消息
	now := time.Now().Unix()
	query := "UPDATE `admin_notification` SET `deleted_at` = ?, `updated_at` = ? WHERE `user_id` = ? AND `read_status` = 1 AND `deleted_at` = 0"
	_, err := r.conn.ExecCtx(ctx, query, now, now, userID)
	return err
}
