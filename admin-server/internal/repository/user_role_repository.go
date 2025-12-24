package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type UserRoleRepository interface {
	ListRoleIDsByUserID(ctx context.Context, userID uint64) ([]uint64, error)
	UpdateUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error
}

type userRoleRepository struct {
	model model.AdminUserRoleModel
	conn  sqlx.SqlConn
}

func NewUserRoleRepository(repo *Repository) UserRoleRepository {
	return &userRoleRepository{model: repo.AdminUserRoleModel, conn: repo.DB}
}

func (r *userRoleRepository) ListRoleIDsByUserID(ctx context.Context, userID uint64) ([]uint64, error) {
	var list []model.AdminUserRole
	query := "select * from admin_user_role where user_id = ?"
	if err := r.conn.QueryRowsCtx(ctx, &list, query, userID); err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(list))
	for _, ur := range list {
		ids = append(ids, ur.RoleId)
	}
	return ids, nil
}

// UpdateUserRoles 更新用户的角色关联（先物理删除旧的，再添加新的）
func (r *userRoleRepository) UpdateUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	// 先物理删除该用户的所有角色关联
	_, err := r.conn.ExecCtx(ctx, "delete from admin_user_role where user_id = ?", userID)
	if err != nil {
		return err
	}

	// 如果有新的角色，添加关联
	if len(roleIDs) > 0 {
		for _, roleID := range roleIDs {
			newUR := &model.AdminUserRole{
				UserId: userID,
				RoleId: roleID,
			}
			_, err := r.model.Insert(ctx, newUR)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
