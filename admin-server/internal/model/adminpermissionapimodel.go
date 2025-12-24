package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminPermissionApiModel = (*customAdminPermissionApiModel)(nil)

type (
	// AdminPermissionApiModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminPermissionApiModel.
	AdminPermissionApiModel interface {
		adminPermissionApiModel
	}

	customAdminPermissionApiModel struct {
		*defaultAdminPermissionApiModel
	}
)

// NewAdminPermissionApiModel returns a model for the database table.
func NewAdminPermissionApiModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminPermissionApiModel {
	return &customAdminPermissionApiModel{
		defaultAdminPermissionApiModel: newAdminPermissionApiModel(conn, c, opts...),
	}
}
