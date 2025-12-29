package repository

import (
	"context"
	"strings"
	"time"

	"postapocgame/admin-server/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// PerformanceLogRepository 性能监控日志仓库
type PerformanceLogRepository interface {
	// FindPage 分页查询性能日志（预留给后续列表接口使用）
	FindPage(ctx context.Context, page, pageSize int64, method, path string, isSlow int64, statusCode int64, startTime, endTime string) ([]model.AdminPerformanceLog, int64, error)
	// Create 创建一条性能日志记录
	Create(ctx context.Context, log *model.AdminPerformanceLog) error
}

type performanceLogRepository struct {
	model model.AdminPerformanceLogModel
	conn  sqlx.SqlConn
}

// NewPerformanceLogRepository 创建性能日志仓库
func NewPerformanceLogRepository(repo *Repository) PerformanceLogRepository {
	return &performanceLogRepository{
		model: repo.AdminPerformanceLogModel,
		conn:  repo.DB,
	}
}

// FindPage 分页查询性能日志
func (r *performanceLogRepository) FindPage(ctx context.Context, page, pageSize int64, method, path string, isSlow int64, statusCode int64, startTime, endTime string) ([]model.AdminPerformanceLog, int64, error) {
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

	if method != "" {
		where = append(where, "method = ?")
		args = append(args, method)
	}
	if path != "" {
		where = append(where, "path LIKE ?")
		args = append(args, "%"+path+"%")
	}
	if isSlow == 0 || isSlow == 1 {
		// 仅当传入 0 或 1 时启用该条件；其他值表示不按是否慢接口过滤
		where = append(where, "is_slow = ?")
		args = append(args, isSlow)
	}
	if statusCode > 0 {
		where = append(where, "status_code = ?")
		args = append(args, statusCode)
	}
	if startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			where = append(where, "created_at >= ?")
			args = append(args, t.Unix())
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			where = append(where, "created_at <= ?")
			args = append(args, t.Unix())
		}
	}

	whereClause := strings.Join(where, " AND ")

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM admin_performance_log WHERE " + whereClause
	if err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 查询列表
	var list []model.AdminPerformanceLog
	query := "SELECT * FROM admin_performance_log WHERE " + whereClause + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	if err := r.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Create 创建一条性能日志记录
func (r *performanceLogRepository) Create(ctx context.Context, log *model.AdminPerformanceLog) error {
	if log == nil {
		return nil
	}
	_, err := r.model.Insert(ctx, log)
	return err
}
