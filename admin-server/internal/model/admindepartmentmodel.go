package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminDepartmentModel = (*customAdminDepartmentModel)(nil)

type (
	// AdminDepartmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminDepartmentModel.
	AdminDepartmentModel interface {
		adminDepartmentModel
	}

	customAdminDepartmentModel struct {
		*defaultAdminDepartmentModel
	}
)

// NewAdminDepartmentModel returns a model for the database table.
func NewAdminDepartmentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminDepartmentModel {
	return &customAdminDepartmentModel{
		defaultAdminDepartmentModel: newAdminDepartmentModel(conn, c, opts...),
	}
}
