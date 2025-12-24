package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminUserModel = (*customAdminUserModel)(nil)

type (
	// AdminUserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminUserModel.
	AdminUserModel interface {
		adminUserModel
	}

	customAdminUserModel struct {
		*defaultAdminUserModel
	}
)

// NewAdminUserModel returns a model for the database table.
func NewAdminUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminUserModel {
	return &customAdminUserModel{
		defaultAdminUserModel: newAdminUserModel(conn, c, opts...),
	}
}
