package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminNotificationModel = (*customAdminNotificationModel)(nil)

type (
	// AdminNotificationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminNotificationModel.
	AdminNotificationModel interface {
		adminNotificationModel
	}

	customAdminNotificationModel struct {
		*defaultAdminNotificationModel
	}
)

// NewAdminNotificationModel returns a model for the database table.
func NewAdminNotificationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AdminNotificationModel {
	return &customAdminNotificationModel{
		defaultAdminNotificationModel: newAdminNotificationModel(conn, c, opts...),
	}
}
