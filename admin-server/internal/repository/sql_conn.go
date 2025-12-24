package repository

import (
	"postapocgame/admin-server/internal/config"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewSQLConn 初始化 go-zero sqlx 连接。
func NewSQLConn(conf config.DatabaseConf) (sqlx.SqlConn, error) {
	if conf.DSN == "" {
		return nil, errors.New("database dsn is empty")
	}
	conn := sqlx.NewMysql(conf.DSN)
	return conn, nil
}
