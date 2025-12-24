package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminMenuModel = (*customAdminMenuModel)(nil)

type (
	// AdminMenuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminMenuModel.
	AdminMenuModel interface {
		adminMenuModel
	}

	customAdminMenuModel struct {
		*defaultAdminMenuModel
	}
)

// NewAdminMenuModel returns a model for the database table.
func NewAdminMenuModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminMenuModel {
	return &customAdminMenuModel{
		defaultAdminMenuModel: newAdminMenuModel(conn, c, opts...),
	}
}
