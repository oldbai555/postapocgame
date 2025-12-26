package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type DictItemRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminDictItem, error)
	FindByTypeID(ctx context.Context, typeID uint64) ([]model.AdminDictItem, error)
	FindPage(ctx context.Context, page, pageSize int64, typeID uint64, label string) ([]model.AdminDictItem, int64, error)
	DeleteByID(ctx context.Context, id uint64) error
	Create(ctx context.Context, dictItem *model.AdminDictItem) error
	Update(ctx context.Context, dictItem *model.AdminDictItem) error
}

type dictItemRepository struct {
	model model.AdminDictItemModel
	conn  sqlx.SqlConn
}

func NewDictItemRepository(repo *Repository) DictItemRepository {
	return &dictItemRepository{model: repo.AdminDictItemModel, conn: repo.DB}
}

func (r *dictItemRepository) FindByID(ctx context.Context, id uint64) (*model.AdminDictItem, error) {
	return r.model.FindOne(ctx, id)
}

func (r *dictItemRepository) FindByTypeID(ctx context.Context, typeID uint64) ([]model.AdminDictItem, error) {
	var list []model.AdminDictItem
	query := "select * from admin_dict_item where deleted_at = 0 and type_id = ? and status = 1 order by sort asc, id asc"
	err := r.conn.QueryRowsCtx(ctx, &list, query, typeID)
	return list, err
}

func (r *dictItemRepository) FindPage(ctx context.Context, page, pageSize int64, typeID uint64, label string) ([]model.AdminDictItem, int64, error) {
	// 目前生成方法不支持复杂过滤，简单复用生成的分页
	return r.model.FindPage(ctx, page, pageSize)
}

func (r *dictItemRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *dictItemRepository) Create(ctx context.Context, dictItem *model.AdminDictItem) error {
	_, err := r.model.Insert(ctx, dictItem)
	return err
}

func (r *dictItemRepository) Update(ctx context.Context, dictItem *model.AdminDictItem) error {
	return r.model.Update(ctx, dictItem)
}
