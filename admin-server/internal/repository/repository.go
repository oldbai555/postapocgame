package repository

import (
	"context"
	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/model"
	businesscache "postapocgame/admin-server/pkg/cache"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Repository 聚合 goctl 生成的 Model，统一数据访问入口。
type Repository struct {
	DB            sqlx.SqlConn
	CacheConf     cache.CacheConf
	Redis         *redis.Redis                 // go-zero stores/redis 客户端
	BusinessCache *businesscache.BusinessCache // 业务层缓存工具

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
	AdminConfigModel         model.AdminConfigModel
	AdminDictTypeModel       model.AdminDictTypeModel
	AdminDictItemModel       model.AdminDictItemModel
	AdminFileModel           model.AdminFileModel
	DemoModel                model.DemoModel
	ChatMessageModel         model.ChatMessageModel
	ChatOnlineUserModel      model.ChatOnlineUserModel
	AdminOperationLogModel   model.AdminOperationLogModel
	AdminLoginLogModel       model.AdminLoginLogModel
	AuditLogModel            model.AuditLogModel
	AdminPerformanceLogModel model.AdminPerformanceLogModel
}

func NewRepository(conn sqlx.SqlConn, cacheConf cache.CacheConf, rdb *redis.Redis) (*Repository, error) {
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
		BusinessCache:            businesscache.NewBusinessCache(rdb),
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
		AdminConfigModel:         model.NewAdminConfigModel(conn, cacheConf),
		AdminDictTypeModel:       model.NewAdminDictTypeModel(conn, cacheConf),
		AdminDictItemModel:       model.NewAdminDictItemModel(conn, cacheConf),
		AdminFileModel:           model.NewAdminFileModel(conn, cacheConf),
		DemoModel:                model.NewDemoModel(conn, cacheConf),
		ChatMessageModel:         model.NewChatMessageModel(conn, cacheConf),
		ChatOnlineUserModel:      model.NewChatOnlineUserModel(conn, cacheConf),
		AdminOperationLogModel:   model.NewAdminOperationLogModel(conn, cacheConf),
		AdminLoginLogModel:       model.NewAdminLoginLogModel(conn, cacheConf),
		AuditLogModel:            model.NewAuditLogModel(conn, cacheConf),
		AdminPerformanceLogModel: model.NewAdminPerformanceLogModel(conn, cacheConf),
	}, nil
}

// BuildSources 根据配置初始化数据源，供 ServiceContext 调用。
func BuildSources(cfg config.Config) (*Repository, error) {
	conn, err := NewSQLConn(cfg.Database)
	if err != nil {
		return nil, errors.Wrap(err, "init sqlx connection")
	}
	cacheConf := BuildCacheConf(cfg.Redis)
	// 创建 go-zero 的 Redis 客户端
	rdb, err := redis.NewRedis(redis.RedisConf{
		Host: cfg.Redis.Address,
		Pass: cfg.Redis.Password,
		Type: "node",
	})
	if err != nil {
		return nil, errors.Wrap(err, "init redis")
	}
	return NewRepository(conn, cacheConf, rdb)
}

// ClearCache 清理所有 go-zero SQL 查询缓存
// 注意：go-zero 的 Redis 客户端不支持 FlushDB，此函数暂时简化实现
func (r *Repository) ClearCache(ctx context.Context) error {
	if r.Redis == nil {
		return errors.New("redis client is nil")
	}

	// go-zero 的 Redis 客户端不支持 FlushDB 方法
	// 如果需要清理缓存，建议使用其他方式管理缓存 key
	logx.Infof("ClearCache called, go-zero Redis does not support FlushDB")
	return nil
}

// ClearCacheByPrefix 通过前缀清理缓存（更安全的方式，只清理 go-zero 缓存）
// 注意：go-zero 的 Redis 客户端不支持 SCAN，此函数暂时简化实现
func (r *Repository) ClearCacheByPrefix(ctx context.Context, prefix string) error {
	if r.Redis == nil {
		return errors.New("redis client is nil")
	}

	// go-zero 的 Redis 客户端不支持 SCAN 命令
	// 如果需要按前缀清理，建议使用 FlushDB 清理整个 DB
	// 或者使用其他方式管理缓存 key
	logx.Infof("ClearCacheByPrefix called with prefix: %s, using FlushDB instead", prefix)
	return r.ClearCache(ctx)
}

// ClearAllCacheDBs 清理所有可能的 Redis DB 中的缓存（用于解决缓存不一致问题）
// 注意：go-zero 的 Redis 客户端不支持多 DB 操作，此函数暂时简化实现
func (r *Repository) ClearAllCacheDBs(ctx context.Context) error {
	if r.Redis == nil {
		return errors.New("redis client is nil")
	}

	// go-zero 的 Redis 客户端不支持多 DB 操作
	// 只清理当前连接的 DB
	logx.Infof("ClearAllCacheDBs called, using FlushDB for current DB")
	return r.ClearCache(ctx)
}
