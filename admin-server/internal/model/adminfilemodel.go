package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminFileModel = (*customAdminFileModel)(nil)

type (
	// AdminFileModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminFileModel.
	AdminFileModel interface {
		adminFileModel
	}

	customAdminFileModel struct {
		*defaultAdminFileModel
	}
)

// NewAdminFileModel returns a model for the database table.
func NewAdminFileModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminFileModel {
	return &customAdminFileModel{
		defaultAdminFileModel: newAdminFileModel(conn, c, opts...),
	}
}
