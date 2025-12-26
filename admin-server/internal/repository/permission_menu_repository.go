package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type PermissionMenuRepository interface {
	ListMenuIDsByPermissionID(ctx context.Context, permissionID uint64) ([]uint64, error)
	UpdatePermissionMenus(ctx context.Context, permissionID uint64, menuIDs []uint64) error
	// ListMenuPermissionCodes 返回「菜单ID -> 绑定的权限编码列表」的完整映射，用于按权限过滤菜单树
	ListMenuPermissionCodes(ctx context.Context) (map[uint64][]string, error)
}

type permissionMenuRepository struct {
	model model.AdminPermissionMenuModel
	conn  sqlx.SqlConn
}

func NewPermissionMenuRepository(repo *Repository) PermissionMenuRepository {
	return &permissionMenuRepository{
		model: repo.AdminPermissionMenuModel,
		conn:  repo.DB,
	}
}

// ListMenuIDsByPermissionID 查询权限关联的菜单ID列表
func (r *permissionMenuRepository) ListMenuIDsByPermissionID(ctx context.Context, permissionID uint64) ([]uint64, error) {
	var list []model.AdminPermissionMenu
	query := "select * from admin_permission_menu where permission_id = ?"
	if err := r.conn.QueryRowsCtx(ctx, &list, query, permissionID); err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(list))
	for _, pm := range list {
		ids = append(ids, pm.MenuId)
	}
	return ids, nil
}

// UpdatePermissionMenus 更新权限的菜单关联（先物理删除旧的，再添加新的）
func (r *permissionMenuRepository) UpdatePermissionMenus(ctx context.Context, permissionID uint64, menuIDs []uint64) error {
	// 物理删除该权限的所有菜单关联
	_, err := r.conn.ExecCtx(ctx, "delete from admin_permission_menu where permission_id = ?", permissionID)
	if err != nil {
		return err
	}

	// 如果有新的菜单，添加关联
	if len(menuIDs) > 0 {
		for _, menuID := range menuIDs {
			newPM := &model.AdminPermissionMenu{
				PermissionId: permissionID,
				MenuId:       menuID,
			}
			_, err := r.model.Insert(ctx, newPM)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ListMenuPermissionCodes 查询所有菜单与权限编码的关联关系
// 用于菜单树过滤：绑定了权限的菜单/按钮，必须命中其中至少一个权限编码才可见
func (r *permissionMenuRepository) ListMenuPermissionCodes(ctx context.Context) (map[uint64][]string, error) {
	type row struct {
		MenuId uint64 `db:"menu_id"`
		Code   string `db:"code"`
	}

	var rows []row
	query := `
SELECT pm.menu_id, p.code
FROM admin_permission_menu pm
JOIN admin_permission p ON pm.permission_id = p.id
WHERE p.deleted_at = 0
`
	if err := r.conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		return nil, err
	}

	result := make(map[uint64][]string)
	for _, r := range rows {
		result[r.MenuId] = append(result[r.MenuId], r.Code)
	}
	return result, nil
}
