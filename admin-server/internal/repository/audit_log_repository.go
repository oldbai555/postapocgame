package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type AuditLogRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AuditLog, error)
	FindPage(ctx context.Context, page, pageSize int64, userId uint64, username, auditType, auditObject, startTime, endTime string) ([]model.AuditLog, int64, error)
	Create(ctx context.Context, log *model.AuditLog) error
}

type auditLogRepository struct {
	model model.AuditLogModel
	conn  sqlx.SqlConn
}

func NewAuditLogRepository(repo *Repository) AuditLogRepository {
	return &auditLogRepository{model: repo.AuditLogModel, conn: repo.DB}
}

func (r *auditLogRepository) FindByID(ctx context.Context, id uint64) (*model.AuditLog, error) {
	return r.model.FindOne(ctx, id)
}

func (r *auditLogRepository) FindPage(ctx context.Context, page, pageSize int64, userId uint64, username, auditType, auditObject, startTime, endTime string) ([]model.AuditLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	where := []string{"deleted_at = 0"}
	args := []interface{}{}

	if userId > 0 {
		where = append(where, "user_id = ?")
		args = append(args, userId)
	}
	if username != "" {
		where = append(where, "username LIKE ?")
		args = append(args, "%"+username+"%")
	}
	if auditType != "" {
		where = append(where, "audit_type = ?")
		args = append(args, auditType)
	}
	if auditObject != "" {
		where = append(where, "audit_object = ?")
		args = append(args, auditObject)
	}
	if startTime != "" {
		// 解析时间字符串为时间戳
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			where = append(where, "created_at >= ?")
			args = append(args, t.Unix())
		}
	}
	if endTime != "" {
		// 解析时间字符串为时间戳
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			where = append(where, "created_at <= ?")
			args = append(args, t.Unix())
		}
	}

	whereClause := strings.Join(where, " AND ")

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_log WHERE %s", whereClause)
	if err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 查询列表
	var list []model.AuditLog
	query := fmt.Sprintf("SELECT * FROM audit_log WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, pageSize, offset)
	if err := r.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *auditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	if log == nil {
		return fmt.Errorf("审计日志数据为空")
	}

	// 设置时间戳
	now := time.Now().Unix()
	if log.CreatedAt == 0 {
		log.CreatedAt = now
	}
	if log.UpdatedAt == 0 {
		log.UpdatedAt = now
	}
	if log.DeletedAt == 0 {
		log.DeletedAt = 0
	}

	_, err := r.model.Insert(ctx, log)
	if err != nil {
		return fmt.Errorf("插入审计日志失败: %w", err)
	}

	return nil
}
