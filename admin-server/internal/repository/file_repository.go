package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type FileRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminFile, error)
	FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminFile, int64, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, file *model.AdminFile) error
	Update(ctx context.Context, file *model.AdminFile) error
}

type fileRepository struct {
	model model.AdminFileModel
	conn  sqlx.SqlConn
}

func NewFileRepository(repo *Repository) FileRepository {
	return &fileRepository{model: repo.AdminFileModel, conn: repo.DB}
}

func (r *fileRepository) FindByID(ctx context.Context, id uint64) (*model.AdminFile, error) {
	return r.model.FindOne(ctx, id)
}

func (r *fileRepository) FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminFile, int64, error) {
	// 目前生成方法不支持复杂过滤，简单复用生成的分页
	return r.model.FindPage(ctx, page, pageSize)
}

func (r *fileRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *fileRepository) Create(ctx context.Context, file *model.AdminFile) error {
	_, err := r.model.Insert(ctx, file)
	return err
}

func (r *fileRepository) Update(ctx context.Context, file *model.AdminFile) error {
	return r.model.Update(ctx, file)
}
