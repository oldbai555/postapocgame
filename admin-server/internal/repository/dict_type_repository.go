package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type DictTypeRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminDictType, error)
	FindByCode(ctx context.Context, code string) (*model.AdminDictType, error)
	FindPage(ctx context.Context, page, pageSize int64, name, code string) ([]model.AdminDictType, int64, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, dictType *model.AdminDictType) error
	Update(ctx context.Context, dictType *model.AdminDictType) error
}

type dictTypeRepository struct {
	model model.AdminDictTypeModel
	conn  sqlx.SqlConn
}

func NewDictTypeRepository(repo *Repository) DictTypeRepository {
	return &dictTypeRepository{model: repo.AdminDictTypeModel, conn: repo.DB}
}

func (r *dictTypeRepository) FindByID(ctx context.Context, id uint64) (*model.AdminDictType, error) {
	return r.model.FindOne(ctx, id)
}

func (r *dictTypeRepository) FindByCode(ctx context.Context, code string) (*model.AdminDictType, error) {
	return r.model.FindOneByCode(ctx, code)
}

func (r *dictTypeRepository) FindPage(ctx context.Context, page, pageSize int64, name, code string) ([]model.AdminDictType, int64, error) {
	// 目前生成方法不支持复杂过滤，简单复用生成的分页
	return r.model.FindPage(ctx, page, pageSize)
}

func (r *dictTypeRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *dictTypeRepository) Create(ctx context.Context, dictType *model.AdminDictType) error {
	_, err := r.model.Insert(ctx, dictType)
	return err
}

func (r *dictTypeRepository) Update(ctx context.Context, dictType *model.AdminDictType) error {
	return r.model.Update(ctx, dictType)
}
