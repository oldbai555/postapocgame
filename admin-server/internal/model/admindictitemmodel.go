package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminDictItemModel = (*customAdminDictItemModel)(nil)

type (
	// AdminDictItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminDictItemModel.
	AdminDictItemModel interface {
		adminDictItemModel
	}

	customAdminDictItemModel struct {
		*defaultAdminDictItemModel
	}
)

// NewAdminDictItemModel returns a model for the database table.
func NewAdminDictItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminDictItemModel {
	return &customAdminDictItemModel{
		defaultAdminDictItemModel: newAdminDictItemModel(conn, c, opts...),
	}
}
