package repository

import (
	"context"
	"errors"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/model"
)

var (
	// ErrRoleNotFound 角色不存在
	ErrRoleNotFound = errors.New("role not found")
)

// RoleRepository 角色数据访问接口
type RoleRepository interface {
	GetRolesByAccountID(ctx context.Context, accountID uint64) ([]*model.Role, error)
	CreateRole(ctx context.Context, accountID uint64, roleName string, job, sex uint32) (*model.Role, error)
	CheckRoleNameExists(ctx context.Context, roleName string) (bool, error)
	GetRoleByID(ctx context.Context, roleID uint64) (*model.Role, error)
}
