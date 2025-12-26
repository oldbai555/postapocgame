package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminConfigModel = (*customAdminConfigModel)(nil)

type (
	// AdminConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminConfigModel.
	AdminConfigModel interface {
		adminConfigModel
	}

	customAdminConfigModel struct {
		*defaultAdminConfigModel
	}
)

// NewAdminConfigModel returns a model for the database table.
func NewAdminConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminConfigModel {
	return &customAdminConfigModel{
		defaultAdminConfigModel: newAdminConfigModel(conn, c, opts...),
	}
}
