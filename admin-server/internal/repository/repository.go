package repository

import (
	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/model"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Repository 聚合 goctl 生成的 Model，统一数据访问入口。
type Repository struct {
	DB        sqlx.SqlConn
	CacheConf cache.CacheConf
	Redis     *redis.Client

	AdminUserModel           model.AdminUserModel
	AdminRoleModel           model.AdminRoleModel
	AdminPermissionModel     model.AdminPermissionModel
	AdminMenuModel           model.AdminMenuModel
	AdminDepartmentModel     model.AdminDepartmentModel
	AdminUserRoleModel       model.AdminUserRoleModel
	AdminRolePermissionModel model.AdminRolePermissionModel
	AdminApiModel            model.AdminApiModel
	AdminPermissionMenuModel model.AdminPermissionMenuModel
	AdminPermissionApiModel  model.AdminPermissionApiModel
}

func NewRepository(conn sqlx.SqlConn, cacheConf cache.CacheConf, rdb *redis.Client) (*Repository, error) {
	if conn == nil {
		return nil, errors.New("repository requires sqlx conn")
	}
	if rdb == nil {
		return nil, errors.New("repository requires redis")
	}
	return &Repository{
		DB:                       conn,
		CacheConf:                cacheConf,
		Redis:                    rdb,
		AdminUserModel:           model.NewAdminUserModel(conn, cacheConf),
		AdminRoleModel:           model.NewAdminRoleModel(conn, cacheConf),
		AdminPermissionModel:     model.NewAdminPermissionModel(conn, cacheConf),
		AdminMenuModel:           model.NewAdminMenuModel(conn, cacheConf),
		AdminDepartmentModel:     model.NewAdminDepartmentModel(conn, cacheConf),
		AdminUserRoleModel:       model.NewAdminUserRoleModel(conn, cacheConf),
		AdminRolePermissionModel: model.NewAdminRolePermissionModel(conn, cacheConf),
		AdminApiModel:            model.NewAdminApiModel(conn, cacheConf),
		AdminPermissionMenuModel: model.NewAdminPermissionMenuModel(conn, cacheConf),
		AdminPermissionApiModel:  model.NewAdminPermissionApiModel(conn, cacheConf),
	}, nil
}

// BuildSources 根据配置初始化数据源，供 ServiceContext 调用。
func BuildSources(cfg config.Config) (*Repository, error) {
	conn, err := NewSQLConn(cfg.Database)
	if err != nil {
		return nil, errors.Wrap(err, "init sqlx connection")
	}
	cacheConf := BuildCacheConf(cfg.Redis)
	rdb, err := NewRedisClient(cfg.Redis)
	if err != nil {
		return nil, errors.Wrap(err, "init redis")
	}
	return NewRepository(conn, cacheConf, rdb)
}
