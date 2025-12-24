package repository

import (
	"postapocgame/admin-server/internal/config"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// BuildCacheConf 将 Redis 配置转换为 go-zero cache 配置。
func BuildCacheConf(redisConf config.RedisConf) cache.CacheConf {
	// 解析 Address，go-zero 需要 host:port 格式
	host := redisConf.Address
	if host == "" {
		host = "127.0.0.1:6379"
	}

	// 如果 Address 包含完整地址，直接使用；否则需要拆分
	// go-zero 的 RedisConf 需要 Host 和 Port 分开，或者使用完整的 Host:Port 格式
	// 根据 go-zero 源码，Host 字段可以接受 "host:port" 格式
	return cache.CacheConf{
		cache.NodeConf{
			RedisConf: redis.RedisConf{
				Host: host, // 直接使用完整地址
				Pass: redisConf.Password,
				Type: "node",
			},
			Weight: 100, // 需要正权重，否则 go-zero 判定为无可用节点
		},
	}
}
