package repository

import (
	"context"
	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/model"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
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
	AdminConfigModel         model.AdminConfigModel
	AdminDictTypeModel       model.AdminDictTypeModel
	AdminDictItemModel       model.AdminDictItemModel
	AdminFileModel           model.AdminFileModel
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
		AdminConfigModel:         model.NewAdminConfigModel(conn, cacheConf),
		AdminDictTypeModel:       model.NewAdminDictTypeModel(conn, cacheConf),
		AdminDictItemModel:       model.NewAdminDictItemModel(conn, cacheConf),
		AdminFileModel:           model.NewAdminFileModel(conn, cacheConf),
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

// ClearCache 清理所有 go-zero SQL 查询缓存
// 注意：go-zero 的缓存可能使用默认 DB (0)，而 Redis 客户端可能使用不同的 DB
// 为了确保清理所有缓存，我们需要清理所有可能的 DB
func (r *Repository) ClearCache(ctx context.Context) error {
	if r.Redis == nil {
		return errors.New("redis client is nil")
	}

	// 获取当前 Redis 客户端使用的 DB
	currentDB := r.Redis.Options().DB

	// 方式1：清理当前 DB（go-zero 缓存可能在这里）
	err := r.Redis.FlushDB(ctx).Err()
	if err != nil {
		return errors.Wrap(err, "flush redis db")
	}

	// 方式2：如果当前 DB 不是 0，也清理 DB 0（go-zero 默认可能使用 DB 0）
	if currentDB != 0 {
		// 创建一个临时客户端连接到 DB 0
		tempClient := redis.NewClient(&redis.Options{
			Addr:     r.Redis.Options().Addr,
			Password: r.Redis.Options().Password,
			DB:       0, // go-zero 缓存可能使用的默认 DB
		})
		defer tempClient.Close()

		if err := tempClient.FlushDB(ctx).Err(); err != nil {
			// 如果清理失败，记录日志但不阻止主流程
			logx.Errorf("清理 Redis DB 0 失败: %v", err)
		}
	}

	return nil
}

// ClearCacheByPrefix 通过前缀清理缓存（更安全的方式，只清理 go-zero 缓存）
func (r *Repository) ClearCacheByPrefix(ctx context.Context, prefix string) error {
	if r.Redis == nil {
		return errors.New("redis client is nil")
	}

	if prefix == "" {
		prefix = "cache:" // 默认清理 go-zero 缓存前缀
	}

	// 使用 SCAN 命令扫描所有匹配的 key
	var cursor uint64
	var deletedCount int64

	for {
		keys, nextCursor, err := r.Redis.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return errors.Wrap(err, "scan redis keys")
		}

		// 批量删除匹配的 key
		if len(keys) > 0 {
			deleted, err := r.Redis.Del(ctx, keys...).Result()
			if err != nil {
				return errors.Wrap(err, "delete redis keys")
			}
			deletedCount += deleted
		}

		cursor = nextCursor
		if cursor == 0 {
			break // 扫描完成
		}
	}

	return nil
}

// ClearAllCacheDBs 清理所有可能的 Redis DB 中的缓存（用于解决缓存不一致问题）
// 清理所有 DB (0-15) 中的 cache: 前缀的 key
func (r *Repository) ClearAllCacheDBs(ctx context.Context) error {
	if r.Redis == nil {
		return errors.New("redis client is nil")
	}

	// 清理所有可能的 DB (0-15) 中的 cache: 前缀的 key
	for db := 0; db < 16; db++ {
		tempClient := redis.NewClient(&redis.Options{
			Addr:     r.Redis.Options().Addr,
			Password: r.Redis.Options().Password,
			DB:       db,
		})

		// 使用 SCAN 清理 cache: 前缀的 key
		var cursor uint64
		var deletedCount int64
		for {
			keys, nextCursor, err := tempClient.Scan(ctx, cursor, "cache:*", 100).Result()
			if err != nil {
				tempClient.Close()
				break
			}

			if len(keys) > 0 {
				deleted, err := tempClient.Del(ctx, keys...).Result()
				if err != nil {
					logx.Errorf("清理 Redis DB %d 的缓存失败: %v", db, err)
				} else {
					deletedCount += deleted
				}
			}

			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}

		if deletedCount > 0 {
			logx.Infof("从 Redis DB %d 清理了 %d 个缓存 key", db, deletedCount)
		}

		tempClient.Close()
	}

	return nil
}
