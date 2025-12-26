package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminDictTypeModel = (*customAdminDictTypeModel)(nil)

type (
	// AdminDictTypeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminDictTypeModel.
	AdminDictTypeModel interface {
		adminDictTypeModel
	}

	customAdminDictTypeModel struct {
		*defaultAdminDictTypeModel
	}
)

// NewAdminDictTypeModel returns a model for the database table.
func NewAdminDictTypeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminDictTypeModel {
	return &customAdminDictTypeModel{
		defaultAdminDictTypeModel: newAdminDictTypeModel(conn, c, opts...),
	}
}
