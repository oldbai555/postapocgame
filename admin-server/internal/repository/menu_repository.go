package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type MenuRepository interface {
	ListAll(ctx context.Context) ([]model.AdminMenu, error)
	FindByID(ctx context.Context, id uint64) (*model.AdminMenu, error)
	Create(ctx context.Context, m *model.AdminMenu) error
	Update(ctx context.Context, m *model.AdminMenu) error
	DeleteByID(ctx context.Context, id uint64) error
}

type menuRepository struct {
	model model.AdminMenuModel
	conn  sqlx.SqlConn
}

func NewMenuRepository(repo *Repository) MenuRepository {
	return &menuRepository{model: repo.AdminMenuModel, conn: repo.DB}
}

func (r *menuRepository) ListAll(ctx context.Context) ([]model.AdminMenu, error) {
	// 直接查询所有未删除的菜单，按 order_num 和 id 排序
	var list []model.AdminMenu
	query := "select id, parent_id, name, path, component, icon, type, order_num, visible, status, created_at, updated_at, deleted_at from admin_menu where deleted_at = 0 order by order_num asc, id asc"
	err := r.conn.QueryRowsCtx(ctx, &list, query)
	return list, err
}

func (r *menuRepository) FindByID(ctx context.Context, id uint64) (*model.AdminMenu, error) {
	return r.model.FindOne(ctx, id)
}

func (r *menuRepository) Create(ctx context.Context, m *model.AdminMenu) error {
	_, err := r.model.Insert(ctx, m)
	return err
}

func (r *menuRepository) Update(ctx context.Context, m *model.AdminMenu) error {
	return r.model.Update(ctx, m)
}

func (r *menuRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}
