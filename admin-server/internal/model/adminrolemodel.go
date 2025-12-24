package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminRoleModel = (*customAdminRoleModel)(nil)

type (
	// AdminRoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminRoleModel.
	AdminRoleModel interface {
		adminRoleModel
	}

	customAdminRoleModel struct {
		*defaultAdminRoleModel
	}
)

// NewAdminRoleModel returns a model for the database table.
func NewAdminRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminRoleModel {
	return &customAdminRoleModel{
		defaultAdminRoleModel: newAdminRoleModel(conn, c, opts...),
	}
}
