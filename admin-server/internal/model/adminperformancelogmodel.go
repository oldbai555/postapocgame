package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminPerformanceLogModel = (*customAdminPerformanceLogModel)(nil)

type (
	// AdminPerformanceLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminPerformanceLogModel.
	AdminPerformanceLogModel interface {
		adminPerformanceLogModel
	}

	customAdminPerformanceLogModel struct {
		*defaultAdminPerformanceLogModel
	}
)

// NewAdminPerformanceLogModel returns a model for the database table.
func NewAdminPerformanceLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminPerformanceLogModel {
	return &customAdminPerformanceLogModel{
		defaultAdminPerformanceLogModel: newAdminPerformanceLogModel(conn, c, opts...),
	}
}
