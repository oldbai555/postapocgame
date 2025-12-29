package repository

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type LoginLogRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminLoginLog, error)
	FindPage(ctx context.Context, page, pageSize int64, userId uint64, username string, status int, startTime, endTime string) ([]model.AdminLoginLog, int64, error)
	Create(ctx context.Context, log *model.AdminLoginLog) error
	// 统计功能
	CountByStatus(ctx context.Context, status int) (int64, error)
	CountToday(ctx context.Context) (int64, error)
	CountTodayByStatus(ctx context.Context, status int) (int64, error)
}

type loginLogRepository struct {
	model model.AdminLoginLogModel
	conn  sqlx.SqlConn
}

func NewLoginLogRepository(repo *Repository) LoginLogRepository {
	return &loginLogRepository{model: repo.AdminLoginLogModel, conn: repo.DB}
}

func (r *loginLogRepository) FindByID(ctx context.Context, id uint64) (*model.AdminLoginLog, error) {
	return r.model.FindOne(ctx, id)
}

func (r *loginLogRepository) FindPage(ctx context.Context, page, pageSize int64, userId uint64, username string, status int, startTime, endTime string) ([]model.AdminLoginLog, int64, error) {
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
	// status < 0 表示不筛选，status >= 0 时添加筛选条件
	// 注意：status 可以是 0（失败）或 1（成功）
	if status >= 0 {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if startTime != "" {
		// 解析时间字符串为时间戳
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			where = append(where, "login_at >= ?")
			args = append(args, t.Unix())
		}
	}
	if endTime != "" {
		// 解析时间字符串为时间戳
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			where = append(where, "login_at <= ?")
			args = append(args, t.Unix())
		}
	}

	whereClause := strings.Join(where, " AND ")

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `admin_login_log` WHERE %s", whereClause)
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	var list []model.AdminLoginLog
	query := fmt.Sprintf("SELECT * FROM `admin_login_log` WHERE %s ORDER BY login_at DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, pageSize, offset)
	err = r.conn.QueryRowsCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *loginLogRepository) Create(ctx context.Context, log *model.AdminLoginLog) error {
	if log == nil {
		return fmt.Errorf("登录日志数据为空")
	}

	// 记录调试信息
	logx.Infof("准备插入登录日志: userId=%d, username=%s, status=%d, message=%s",
		log.UserId, log.Username, log.Status, log.Message)

	result, err := r.model.Insert(ctx, log)
	if err != nil {
		logx.Errorf("插入登录日志失败: userId=%d, username=%s, error: %v",
			log.UserId, log.Username, err)
		return fmt.Errorf("插入登录日志失败: %w", err)
	}

	// 获取插入的 ID（用于调试）
	if id, err := result.LastInsertId(); err == nil {
		logx.Infof("成功插入登录日志: id=%d, userId=%d, username=%s", id, log.UserId, log.Username)
	}
	return nil
}

func (r *loginLogRepository) CountByStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	var query string
	var err error
	if status < 0 {
		// status < 0 表示查询所有状态
		query = "SELECT COUNT(*) FROM `admin_login_log` WHERE deleted_at = 0"
		err = r.conn.QueryRowCtx(ctx, &count, query)
	} else {
		query = "SELECT COUNT(*) FROM `admin_login_log` WHERE deleted_at = 0 AND status = ?"
		err = r.conn.QueryRowCtx(ctx, &count, query, status)
	}
	return count, err
}

func (r *loginLogRepository) CountToday(ctx context.Context) (int64, error) {
	var count int64
	// 获取今天的开始时间戳（00:00:00）
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	query := "SELECT COUNT(*) FROM `admin_login_log` WHERE deleted_at = 0 AND login_at >= ?"
	err := r.conn.QueryRowCtx(ctx, &count, query, todayStart)
	return count, err
}

func (r *loginLogRepository) CountTodayByStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	// 获取今天的开始时间戳（00:00:00）
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	query := "SELECT COUNT(*) FROM `admin_login_log` WHERE deleted_at = 0 AND status = ? AND login_at >= ?"
	err := r.conn.QueryRowCtx(ctx, &count, query, status, todayStart)
	return count, err
}
