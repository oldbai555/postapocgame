package repository

import (
	"database/sql"
	"reflect"
	"time"

	"postapocgame/admin-server/internal/config"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NewSQLConn 初始化 go-zero sqlx 连接，并配置连接池参数。
func NewSQLConn(conf config.DatabaseConf) (sqlx.SqlConn, error) {
	if conf.DSN == "" {
		return nil, errors.New("database dsn is empty")
	}
	conn := sqlx.NewMysql(conf.DSN)

	// 配置连接池参数
	// go-zero sqlx 内部使用 database/sql，通过反射获取底层 *sql.DB
	if err := configureConnectionPool(conn, conf); err != nil {
		logx.Errorf("配置数据库连接池失败: %v", err)
		// 不返回错误，继续使用默认连接池配置
	}

	return conn, nil
}

// configureConnectionPool 通过反射获取底层 *sql.DB 并配置连接池
func configureConnectionPool(conn sqlx.SqlConn, conf config.DatabaseConf) error {
	// 使用反射获取底层 *sql.DB
	// go-zero sqlx.MysqlConn 内部有 conn 字段，类型为 *sql.DB
	connValue := reflect.ValueOf(conn)
	if connValue.Kind() == reflect.Ptr {
		connValue = connValue.Elem()
	}

	// 查找 conn 字段（go-zero sqlx.MysqlConn 的内部字段）
	connField := connValue.FieldByName("conn")
	if !connField.IsValid() {
		// 如果找不到 conn 字段，尝试其他可能的字段名
		connField = connValue.FieldByName("db")
	}
	if !connField.IsValid() {
		return errors.New("无法获取底层 *sql.DB，连接池配置将使用默认值")
	}

	// 获取 *sql.DB
	db, ok := connField.Interface().(*sql.DB)
	if !ok {
		return errors.New("底层连接不是 *sql.DB 类型")
	}

	// 设置连接池参数
	maxOpen := conf.MaxOpen
	if maxOpen <= 0 {
		maxOpen = 20 // 默认值
	}
	maxIdle := conf.MaxIdle
	if maxIdle <= 0 {
		maxIdle = 10 // 默认值
	}
	connMaxLifetime := time.Duration(conf.ConnMaxLifetime) * time.Second
	if conf.ConnMaxLifetime <= 0 {
		connMaxLifetime = 5 * time.Minute // 默认 5 分钟
	}
	connMaxIdleTime := time.Duration(conf.ConnMaxIdleTime) * time.Second
	if conf.ConnMaxIdleTime <= 0 {
		connMaxIdleTime = 10 * time.Minute // 默认 10 分钟
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	logx.Infof("数据库连接池已配置: MaxOpen=%d, MaxIdle=%d, ConnMaxLifetime=%v, ConnMaxIdleTime=%v",
		maxOpen, maxIdle, connMaxLifetime, connMaxIdleTime)

	return nil
}
