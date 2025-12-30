package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChatUserModel = (*customChatUserModel)(nil)

type (
	// ChatUserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatUserModel.
	ChatUserModel interface {
		chatUserModel
	}

	customChatUserModel struct {
		*defaultChatUserModel
	}
)

// NewChatUserModel returns a model for the database table.
func NewChatUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ChatUserModel {
	return &customChatUserModel{
		defaultChatUserModel: newChatUserModel(conn, c, opts...),
	}
}
