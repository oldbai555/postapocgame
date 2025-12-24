package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type RoleRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminRole, error)
	FindByCode(ctx context.Context, code string) (*model.AdminRole, error)
	FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminRole, int64, error)
	FindChunk(ctx context.Context, limit int64, lastId uint64) ([]model.AdminRole, uint64, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, role *model.AdminRole) error
	Update(ctx context.Context, role *model.AdminRole) error
}

type roleRepository struct {
	model model.AdminRoleModel
	conn  sqlx.SqlConn
}

func NewRoleRepository(repo *Repository) RoleRepository {
	return &roleRepository{model: repo.AdminRoleModel, conn: repo.DB}
}

func (r *roleRepository) FindByID(ctx context.Context, id uint64) (*model.AdminRole, error) {
	return r.model.FindOne(ctx, id)
}

func (r *roleRepository) FindByCode(ctx context.Context, code string) (*model.AdminRole, error) {
	return r.model.FindOneByCode(ctx, code)
}

// FindPage 分页查询角色列表（符合新规范）
func (r *roleRepository) FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminRole, int64, error) {
	// 目前生成方法不支持模糊过滤，简单复用生成的分页
	return r.model.FindPage(ctx, page, pageSize)
}

// FindChunk 分片查询角色列表（基于lastId，适用于大数据量分批处理）
func (r *roleRepository) FindChunk(ctx context.Context, limit int64, lastId uint64) ([]model.AdminRole, uint64, error) {
	return r.model.FindChunk(ctx, limit, lastId)
}

func (r *roleRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *roleRepository) Create(ctx context.Context, role *model.AdminRole) error {
	_, err := r.model.Insert(ctx, role)
	return err
}

func (r *roleRepository) Update(ctx context.Context, role *model.AdminRole) error {
	return r.model.Update(ctx, role)
}
