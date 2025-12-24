package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type DepartmentRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminDepartment, error)
	ListAll(ctx context.Context) ([]model.AdminDepartment, error)
	ListChildren(ctx context.Context, parentID uint64) ([]model.AdminDepartment, error)
	Create(ctx context.Context, dept *model.AdminDepartment) error
	Update(ctx context.Context, dept *model.AdminDepartment) error
	DeleteByID(ctx context.Context, id uint64) error
}

type departmentRepository struct {
	model model.AdminDepartmentModel
	conn  sqlx.SqlConn
}

func NewDepartmentRepository(repo *Repository) DepartmentRepository {
	return &departmentRepository{model: repo.AdminDepartmentModel, conn: repo.DB}
}

func (r *departmentRepository) FindByID(ctx context.Context, id uint64) (*model.AdminDepartment, error) {
	return r.model.FindOne(ctx, id)
}

func (r *departmentRepository) ListAll(ctx context.Context) ([]model.AdminDepartment, error) {
	list, _, err := r.model.FindPage(ctx, 1, 10000)
	return list, err
}

func (r *departmentRepository) ListChildren(ctx context.Context, parentID uint64) ([]model.AdminDepartment, error) {
	var list []model.AdminDepartment
	query := "select * from admin_department where deleted_at = 0 and parent_id = ? order by order_num asc, id asc"
	err := r.conn.QueryRowsCtx(ctx, &list, query, parentID)
	return list, err
}

func (r *departmentRepository) Create(ctx context.Context, dept *model.AdminDepartment) error {
	_, err := r.model.Insert(ctx, dept)
	return err
}

func (r *departmentRepository) Update(ctx context.Context, dept *model.AdminDepartment) error {
	return r.model.Update(ctx, dept)
}

func (r *departmentRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}
