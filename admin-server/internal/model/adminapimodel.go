package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminApiModel = (*customAdminApiModel)(nil)

type (
	// AdminApiModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminApiModel.
	AdminApiModel interface {
		adminApiModel
	}

	customAdminApiModel struct {
		*defaultAdminApiModel
	}
)

// NewAdminApiModel returns a model for the database table.
func NewAdminApiModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminApiModel {
	return &customAdminApiModel{
		defaultAdminApiModel: newAdminApiModel(conn, c, opts...),
	}
}
