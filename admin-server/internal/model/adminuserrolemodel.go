package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminUserRoleModel = (*customAdminUserRoleModel)(nil)

type (
	// AdminUserRoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminUserRoleModel.
	AdminUserRoleModel interface {
		adminUserRoleModel
	}

	customAdminUserRoleModel struct {
		*defaultAdminUserRoleModel
	}
)

// NewAdminUserRoleModel returns a model for the database table.
func NewAdminUserRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminUserRoleModel {
	return &customAdminUserRoleModel{
		defaultAdminUserRoleModel: newAdminUserRoleModel(conn, c, opts...),
	}
}
