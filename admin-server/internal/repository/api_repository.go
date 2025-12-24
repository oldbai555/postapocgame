package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type ApiRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminApi, error)
	FindByMethodAndPath(ctx context.Context, method, path string) (*model.AdminApi, error)
	FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminApi, int64, error)
	Create(ctx context.Context, api *model.AdminApi) error
	Update(ctx context.Context, api *model.AdminApi) error
	DeleteByID(ctx context.Context, id uint64) error
}

type apiRepository struct {
	model model.AdminApiModel
	conn  sqlx.SqlConn
}

func NewApiRepository(repo *Repository) ApiRepository {
	return &apiRepository{model: repo.AdminApiModel, conn: repo.DB}
}

func (r *apiRepository) FindByID(ctx context.Context, id uint64) (*model.AdminApi, error) {
	return r.model.FindOne(ctx, id)
}

func (r *apiRepository) FindByMethodAndPath(ctx context.Context, method, path string) (*model.AdminApi, error) {
	return r.model.FindOneByMethodPath(ctx, method, path)
}

func (r *apiRepository) FindPage(ctx context.Context, page, pageSize int64, name string) ([]model.AdminApi, int64, error) {
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

	var (
		list  []model.AdminApi
		total int64
	)

	if name == "" {
		return r.model.FindPage(ctx, page, pageSize)
	}

	// 带名称模糊筛选的自定义查询
	countQuery := "select count(*) from admin_api where deleted_at = 0 and name like ?"
	if err := r.conn.QueryRowCtx(ctx, &total, countQuery, "%"+name+"%"); err != nil {
		return nil, 0, err
	}
	query := "select * from admin_api where deleted_at = 0 and name like ? order by id desc limit ? offset ?"
	if err := r.conn.QueryRowsCtx(ctx, &list, query, "%"+name+"%", pageSize, offset); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *apiRepository) Create(ctx context.Context, api *model.AdminApi) error {
	_, err := r.model.Insert(ctx, api)
	return err
}

func (r *apiRepository) Update(ctx context.Context, api *model.AdminApi) error {
	return r.model.Update(ctx, api)
}

func (r *apiRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}
