package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type RolePermissionRepository interface {
	ListPermissionIDsByRoleID(ctx context.Context, roleID uint64) ([]uint64, error)
	UpdateRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error
}

type rolePermissionRepository struct {
	model model.AdminRolePermissionModel
	conn  sqlx.SqlConn
}

func NewRolePermissionRepository(repo *Repository) RolePermissionRepository {
	return &rolePermissionRepository{
		model: repo.AdminRolePermissionModel,
		conn:  repo.DB,
	}
}

// ListPermissionIDsByRoleID 查询角色拥有的权限ID列表
func (r *rolePermissionRepository) ListPermissionIDsByRoleID(ctx context.Context, roleID uint64) ([]uint64, error) {
	var list []model.AdminRolePermission
	query := "select * from admin_role_permission where role_id = ?"
	if err := r.conn.QueryRowsCtx(ctx, &list, query, roleID); err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(list))
	for _, rp := range list {
		ids = append(ids, rp.PermissionId)
	}
	return ids, nil
}

// UpdateRolePermissions 更新角色的权限关联（先物理删除旧的，再添加新的）
func (r *rolePermissionRepository) UpdateRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	// 先物理删除该角色的所有权限关联
	_, err := r.conn.ExecCtx(ctx, "delete from admin_role_permission where role_id = ?", roleID)
	if err != nil {
		return err
	}

	// 如果有新的权限，添加关联
	if len(permissionIDs) > 0 {
		for _, permID := range permissionIDs {
			newRP := &model.AdminRolePermission{
				RoleId:       roleID,
				PermissionId: permID,
			}
			_, err := r.model.Insert(ctx, newRP)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
