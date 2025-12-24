package repository

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/config"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient 初始化 redis 客户端并进行连通性校验。
func NewRedisClient(conf config.RedisConf) (*redis.Client, error) {
	if conf.Address == "" {
		return nil, errors.New("redis address is empty")
	}

	// 设置超时时间，默认 5 秒
	dialTimeout := 5 * time.Second
	if conf.DialTimeout > 0 {
		dialTimeout = time.Duration(conf.DialTimeout) * time.Second
	}

	readTimeout := 3 * time.Second
	writeTimeout := 3 * time.Second
	if conf.Timeout > 0 {
		readTimeout = time.Duration(conf.Timeout) * time.Second
		writeTimeout = time.Duration(conf.Timeout) * time.Second
	}

	client := redis.NewClient(&redis.Options{
		Addr:         conf.Address,
		Password:     conf.Password,
		DB:           conf.DB,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// 使用更长的超时时间进行连通性校验
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout+2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, errors.Wrapf(err, "ping redis at %s", conf.Address)
	}

	return client, nil
}
