package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChatOnlineUserModel = (*customChatOnlineUserModel)(nil)

type (
	// ChatOnlineUserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatOnlineUserModel.
	ChatOnlineUserModel interface {
		chatOnlineUserModel
	}

	customChatOnlineUserModel struct {
		*defaultChatOnlineUserModel
	}
)

// NewChatOnlineUserModel returns a model for the database table.
func NewChatOnlineUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ChatOnlineUserModel {
	return &customChatOnlineUserModel{
		defaultChatOnlineUserModel: newChatOnlineUserModel(conn, c, opts...),
	}
}
