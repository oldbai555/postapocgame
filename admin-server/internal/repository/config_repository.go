package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type ConfigRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminConfig, error)
	FindByKey(ctx context.Context, key string) (*model.AdminConfig, error)
	FindPage(ctx context.Context, page, pageSize int64, group, key string) ([]model.AdminConfig, int64, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, config *model.AdminConfig) error
	Update(ctx context.Context, config *model.AdminConfig) error
}

type configRepository struct {
	model model.AdminConfigModel
	conn  sqlx.SqlConn
}

func NewConfigRepository(repo *Repository) ConfigRepository {
	return &configRepository{model: repo.AdminConfigModel, conn: repo.DB}
}

func (r *configRepository) FindByID(ctx context.Context, id uint64) (*model.AdminConfig, error) {
	return r.model.FindOne(ctx, id)
}

func (r *configRepository) FindByKey(ctx context.Context, key string) (*model.AdminConfig, error) {
	return r.model.FindOneByKey(ctx, key)
}

func (r *configRepository) FindPage(ctx context.Context, page, pageSize int64, group, key string) ([]model.AdminConfig, int64, error) {
	// 目前生成方法不支持复杂过滤，简单复用生成的分页
	return r.model.FindPage(ctx, page, pageSize)
}

func (r *configRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *configRepository) Create(ctx context.Context, config *model.AdminConfig) error {
	_, err := r.model.Insert(ctx, config)
	return err
}

func (r *configRepository) Update(ctx context.Context, config *model.AdminConfig) error {
	return r.model.Update(ctx, config)
}
