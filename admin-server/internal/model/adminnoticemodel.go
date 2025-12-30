package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminNoticeModel = (*customAdminNoticeModel)(nil)

type (
	// AdminNoticeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminNoticeModel.
	AdminNoticeModel interface {
		adminNoticeModel
	}

	customAdminNoticeModel struct {
		*defaultAdminNoticeModel
	}
)

// NewAdminNoticeModel returns a model for the database table.
func NewAdminNoticeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminNoticeModel {
	return &customAdminNoticeModel{
		defaultAdminNoticeModel: newAdminNoticeModel(conn, c, opts...),
	}
}
