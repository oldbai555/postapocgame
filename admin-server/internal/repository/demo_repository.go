package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type DemoRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Demo, error)
	FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.Demo, int64, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, demo *model.Demo) error
	Update(ctx context.Context, demo *model.Demo) error
}

type demoRepository struct {
	model model.DemoModel
	conn  sqlx.SqlConn
}

func NewDemoRepository(repo *Repository) DemoRepository {
	return &demoRepository{model: repo.DemoModel, conn: repo.DB}
}

func (r *demoRepository) FindByID(ctx context.Context, id uint64) (*model.Demo, error) {
	return r.model.FindOne(ctx, id)
}

func (r *demoRepository) FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.Demo, int64, error) {
	// 目前生成方法不支持复杂过滤，简单复用生成的分页
	return r.model.FindPage(ctx, page, pageSize)
}

func (r *demoRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *demoRepository) Create(ctx context.Context, demo *model.Demo) error {
	_, err := r.model.Insert(ctx, demo)
	return err
}

func (r *demoRepository) Update(ctx context.Context, demo *model.Demo) error {
	return r.model.Update(ctx, demo)
}
