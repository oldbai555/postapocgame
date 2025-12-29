package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type OperationLogRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminOperationLog, error)
	FindPage(ctx context.Context, page, pageSize int64, userId uint64, username, operationType, operationObject, method, startTime, endTime string) ([]model.AdminOperationLog, int64, error)
	Create(ctx context.Context, log *model.AdminOperationLog) error
	// 批量创建（用于异步写入）
	BatchCreate(ctx context.Context, logs []*model.AdminOperationLog) error
}

type operationLogRepository struct {
	model model.AdminOperationLogModel
	conn  sqlx.SqlConn
}

func NewOperationLogRepository(repo *Repository) OperationLogRepository {
	return &operationLogRepository{model: repo.AdminOperationLogModel, conn: repo.DB}
}

func (r *operationLogRepository) FindByID(ctx context.Context, id uint64) (*model.AdminOperationLog, error) {
	return r.model.FindOne(ctx, id)
}

func (r *operationLogRepository) FindPage(ctx context.Context, page, pageSize int64, userId uint64, username, operationType, operationObject, method, startTime, endTime string) ([]model.AdminOperationLog, int64, error) {
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
	if operationType != "" {
		where = append(where, "operation_type = ?")
		args = append(args, operationType)
	}
	if operationObject != "" {
		where = append(where, "operation_object = ?")
		args = append(args, operationObject)
	}
	if method != "" {
		where = append(where, "method = ?")
		args = append(args, method)
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_operation_log WHERE %s", whereClause)
	if err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 查询列表
	var list []model.AdminOperationLog
	query := fmt.Sprintf("SELECT * FROM admin_operation_log WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, pageSize, offset)
	if err := r.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *operationLogRepository) Create(ctx context.Context, log *model.AdminOperationLog) error {
	// go-zero 生成的 Model 会自动处理 created_at 和 updated_at
	_, err := r.model.Insert(ctx, log)
	return err
}

func (r *operationLogRepository) BatchCreate(ctx context.Context, logs []*model.AdminOperationLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 批量插入（使用事务或循环插入）
	for _, log := range logs {
		_, err := r.model.Insert(ctx, log)
		if err != nil {
			return err
		}
	}
	return nil
}
