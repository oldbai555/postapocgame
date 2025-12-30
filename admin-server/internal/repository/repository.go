package repository

import (
	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/model"
	businesscache "postapocgame/admin-server/pkg/cache"

	"github.com/pkg/errors"
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
	ChatModel                model.ChatModel
	ChatUserModel            model.ChatUserModel
	ChatMessageModel         model.ChatMessageModel
	AdminOperationLogModel   model.AdminOperationLogModel
	AdminLoginLogModel       model.AdminLoginLogModel
	AuditLogModel            model.AuditLogModel
	AdminPerformanceLogModel model.AdminPerformanceLogModel
	AdminNoticeModel         model.AdminNoticeModel
	AdminNotificationModel   model.AdminNotificationModel
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
		ChatModel:                model.NewChatModel(conn, cacheConf),
		ChatUserModel:            model.NewChatUserModel(conn, cacheConf),
		ChatMessageModel:         model.NewChatMessageModel(conn, cacheConf),
		AdminOperationLogModel:   model.NewAdminOperationLogModel(conn, cacheConf),
		AdminLoginLogModel:       model.NewAdminLoginLogModel(conn, cacheConf),
		AuditLogModel:            model.NewAuditLogModel(conn, cacheConf),
		AdminPerformanceLogModel: model.NewAdminPerformanceLogModel(conn, cacheConf),
		AdminNoticeModel:         model.NewAdminNoticeModel(conn, cacheConf),
		AdminNotificationModel:   model.NewAdminNotificationModel(conn, cacheConf),
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
