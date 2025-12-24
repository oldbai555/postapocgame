package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type PermissionApiRepository interface {
	ListApiIDsByPermissionID(ctx context.Context, permissionID uint64) ([]uint64, error)
	ListPermissionIDsByApiID(ctx context.Context, apiID uint64) ([]uint64, error)
	UpdatePermissionApis(ctx context.Context, permissionID uint64, apiIDs []uint64) error
}

type permissionApiRepository struct {
	model model.AdminPermissionApiModel
	conn  sqlx.SqlConn
}

func NewPermissionApiRepository(repo *Repository) PermissionApiRepository {
	return &permissionApiRepository{
		model: repo.AdminPermissionApiModel,
		conn:  repo.DB,
	}
}

// ListApiIDsByPermissionID 查询权限关联的接口ID列表
func (r *permissionApiRepository) ListApiIDsByPermissionID(ctx context.Context, permissionID uint64) ([]uint64, error) {
	var list []model.AdminPermissionApi
	query := "select * from admin_permission_api where permission_id = ?"
	if err := r.conn.QueryRowsCtx(ctx, &list, query, permissionID); err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(list))
	for _, pa := range list {
		ids = append(ids, pa.ApiId)
	}
	return ids, nil
}

// ListPermissionIDsByApiID 查询接口关联的权限ID列表
func (r *permissionApiRepository) ListPermissionIDsByApiID(ctx context.Context, apiID uint64) ([]uint64, error) {
	var list []model.AdminPermissionApi
	query := "select * from admin_permission_api where api_id = ?"
	if err := r.conn.QueryRowsCtx(ctx, &list, query, apiID); err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(list))
	for _, pa := range list {
		ids = append(ids, pa.PermissionId)
	}
	return ids, nil
}

// UpdatePermissionApis 更新权限的接口关联（先物理删除旧的，再添加新的）
func (r *permissionApiRepository) UpdatePermissionApis(ctx context.Context, permissionID uint64, apiIDs []uint64) error {
	// 物理删除该权限的所有接口关联
	_, err := r.conn.ExecCtx(ctx, "delete from admin_permission_api where permission_id = ?", permissionID)
	if err != nil {
		return err
	}

	// 如果有新的接口，添加关联
	if len(apiIDs) > 0 {
		for _, apiID := range apiIDs {
			newPA := &model.AdminPermissionApi{
				PermissionId: permissionID,
				ApiId:        apiID,
			}
			_, err := r.model.Insert(ctx, newPA)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
