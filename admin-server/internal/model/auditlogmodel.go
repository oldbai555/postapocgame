package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuditLogModel = (*customAuditLogModel)(nil)

type (
	// AuditLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuditLogModel.
	AuditLogModel interface {
		auditLogModel
	}

	customAuditLogModel struct {
		*defaultAuditLogModel
	}
)

// NewAuditLogModel returns a model for the database table.
func NewAuditLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuditLogModel {
	return &customAuditLogModel{
		defaultAuditLogModel: newAuditLogModel(conn, c, opts...),
	}
}
