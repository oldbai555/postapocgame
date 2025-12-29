package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminOperationLogModel = (*customAdminOperationLogModel)(nil)

type (
	// AdminOperationLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminOperationLogModel.
	AdminOperationLogModel interface {
		adminOperationLogModel
	}

	customAdminOperationLogModel struct {
		*defaultAdminOperationLogModel
	}
)

// NewAdminOperationLogModel returns a model for the database table.
func NewAdminOperationLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminOperationLogModel {
	return &customAdminOperationLogModel{
		defaultAdminOperationLogModel: newAdminOperationLogModel(conn, c, opts...),
	}
}
