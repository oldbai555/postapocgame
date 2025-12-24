package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminPermissionModel = (*customAdminPermissionModel)(nil)

type (
	// AdminPermissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminPermissionModel.
	AdminPermissionModel interface {
		adminPermissionModel
	}

	customAdminPermissionModel struct {
		*defaultAdminPermissionModel
	}
)

// NewAdminPermissionModel returns a model for the database table.
func NewAdminPermissionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminPermissionModel {
	return &customAdminPermissionModel{
		defaultAdminPermissionModel: newAdminPermissionModel(conn, c, opts...),
	}
}
