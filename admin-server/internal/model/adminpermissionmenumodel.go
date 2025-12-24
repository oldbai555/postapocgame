package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminPermissionMenuModel = (*customAdminPermissionMenuModel)(nil)

type (
	// AdminPermissionMenuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminPermissionMenuModel.
	AdminPermissionMenuModel interface {
		adminPermissionMenuModel
	}

	customAdminPermissionMenuModel struct {
		*defaultAdminPermissionMenuModel
	}
)

// NewAdminPermissionMenuModel returns a model for the database table.
func NewAdminPermissionMenuModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminPermissionMenuModel {
	return &customAdminPermissionMenuModel{
		defaultAdminPermissionMenuModel: newAdminPermissionMenuModel(conn, c, opts...),
	}
}
