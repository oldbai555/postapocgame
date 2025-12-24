package repository

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type PermissionRepository interface {
	ListByRoleIDs(ctx context.Context, roleIDs []uint64) ([]model.AdminPermission, error)
	FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminPermission, int64, error)
	FindChunk(ctx context.Context, limit int64, lastId uint64) ([]model.AdminPermission, uint64, error)
	FindByID(ctx context.Context, id uint64) (*model.AdminPermission, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, p *model.AdminPermission) error
	Update(ctx context.Context, p *model.AdminPermission) error
}

type permissionRepository struct {
	model model.AdminPermissionModel
	conn  sqlx.SqlConn
}

func NewPermissionRepository(repo *Repository) PermissionRepository {
	return &permissionRepository{model: repo.AdminPermissionModel, conn: repo.DB}
}

// ListByRoleIDs 简单实现：通过关联表查询角色拥有的权限列表。
func (r *permissionRepository) ListByRoleIDs(ctx context.Context, roleIDs []uint64) ([]model.AdminPermission, error) {
	if len(roleIDs) == 0 {
		return []model.AdminPermission{}, nil
	}

	var list []model.AdminPermission
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(roleIDs)), ",")
	query := "select p.* from admin_permission p join admin_role_permission arp on arp.permission_id = p.id where p.deleted_at = 0 and arp.role_id in (" + placeholders + ")"
	args := make([]interface{}, 0, len(roleIDs))
	for _, id := range roleIDs {
		args = append(args, id)
	}
	err := r.conn.QueryRowsCtx(ctx, &list, query, args...)
	return list, err
}

// FindPage 分页查询权限列表（符合新规范）
func (r *permissionRepository) FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminPermission, int64, error) {
	return r.model.FindPage(ctx, page, pageSize)
}

// FindChunk 分片查询权限列表（基于lastId，适用于大数据量分批处理）
func (r *permissionRepository) FindChunk(ctx context.Context, limit int64, lastId uint64) ([]model.AdminPermission, uint64, error) {
	return r.model.FindChunk(ctx, limit, lastId)
}

func (r *permissionRepository) FindByID(ctx context.Context, id uint64) (*model.AdminPermission, error) {
	return r.model.FindOne(ctx, id)
}

func (r *permissionRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *permissionRepository) Create(ctx context.Context, p *model.AdminPermission) error {
	_, err := r.model.Insert(ctx, p)
	return err
}

func (r *permissionRepository) Update(ctx context.Context, p *model.AdminPermission) error {
	return r.model.Update(ctx, p)
}
