package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// 缓存 key 前缀
	CacheKeyUserPermissions = "cache:user:permissions:" // 用户权限列表
	CacheKeyUserMenuTree    = "cache:user:menu_tree:"   // 用户菜单树
	CacheKeyMenuTree        = "cache:menu:tree"         // 完整菜单树
	CacheKeyDictItems       = "cache:dict:items:"       // 字典项列表（按 code）
	CacheKeyDictItemsByType = "cache:dict:items:type:"  // 字典项列表（按 type_id）
	CacheKeyConfigList      = "cache:config:list"       // 配置列表
	CacheKeyConfigKey       = "cache:config:key:"       // 单个配置（按 key）

	// 缓存过期时间（秒）
	CacheExpireUserPermissions = 30 * 60 // 30分钟
	CacheExpireUserMenuTree    = 30 * 60 // 30分钟
	CacheExpireMenuTree        = 30 * 60 // 30分钟
	CacheExpireDictItems       = 60 * 60 // 1小时
	CacheExpireConfigList      = 10 * 60 // 10分钟
	CacheExpireConfigKey       = 10 * 60 // 10分钟

	// 缓存过期时间随机范围（秒），用于防止缓存雪崩
	CacheExpireRandomRange = 300 // 5分钟
)

// BusinessCache 业务层缓存工具
type BusinessCache struct {
	redis *redis.Redis
}

// NewBusinessCache 创建业务层缓存工具
func NewBusinessCache(rdb *redis.Redis) *BusinessCache {
	if rdb == nil {
		panic("redis client is nil")
	}
	return &BusinessCache{redis: rdb}
}

// Get 获取缓存值
func (c *BusinessCache) Get(ctx context.Context, key string, v interface{}) error {
	val, err := c.redis.Get(key)
	if err != nil {
		// go-zero Redis Get 返回空字符串表示 key 不存在
		if val == "" {
			return ErrCacheMiss
		}
		return errors.Wrap(err, "redis get")
	}

	// 如果 val 为空字符串，也表示缓存不存在
	if val == "" {
		return ErrCacheMiss
	}

	if v == nil {
		return nil
	}

	if err := json.Unmarshal([]byte(val), v); err != nil {
		return errors.Wrap(err, "json unmarshal")
	}

	return nil
}

// Set 设置缓存值
func (c *BusinessCache) Set(ctx context.Context, key string, v interface{}, expireSeconds int) error {
	data, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "json marshal")
	}

	// 添加随机值防止缓存雪崩
	expireSeconds = c.addRandomExpire(expireSeconds)

	if err := c.redis.Setex(key, string(data), expireSeconds); err != nil {
		return errors.Wrap(err, "redis setex")
	}

	return nil
}

// Delete 删除缓存
func (c *BusinessCache) Delete(ctx context.Context, key string) error {
	_, err := c.redis.Del(key)
	if err != nil {
		return errors.Wrap(err, "redis del")
	}
	return nil
}

// DeleteByPrefix 按前缀删除缓存（注意：go-zero Redis 不支持 SCAN，此方法需要手动管理 key）
func (c *BusinessCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	// go-zero Redis 不支持 SCAN，需要手动管理 key
	// 这里只记录日志，实际删除需要在业务层手动管理
	logx.Infof("DeleteByPrefix called with prefix: %s, go-zero Redis does not support SCAN", prefix)
	return nil
}

// GetOrSet 获取缓存，如果不存在则设置
func (c *BusinessCache) GetOrSet(ctx context.Context, key string, v interface{}, expireSeconds int, fn func() (interface{}, error)) error {
	// 先尝试获取缓存
	err := c.Get(ctx, key, v)
	if err == nil {
		return nil
	}
	if err != ErrCacheMiss {
		return err
	}

	// 缓存不存在，执行回调函数获取数据
	data, err := fn()
	if err != nil {
		return err
	}

	// 设置缓存
	if err := c.Set(ctx, key, data, expireSeconds); err != nil {
		logx.Errorf("设置缓存失败: key=%s, error=%v", key, err)
		// 不返回错误，继续返回数据
	}

	// 将数据赋值给 v
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "json marshal")
	}
	if err := json.Unmarshal(dataBytes, v); err != nil {
		return errors.Wrap(err, "json unmarshal")
	}

	return nil
}

// addRandomExpire 添加随机过期时间，防止缓存雪崩
func (c *BusinessCache) addRandomExpire(expireSeconds int) int {
	if expireSeconds <= CacheExpireRandomRange {
		return expireSeconds
	}
	random := rand.Intn(CacheExpireRandomRange)
	return expireSeconds + random
}

// 错误定义
var (
	ErrCacheMiss = errors.New("cache miss")
)

// 用户权限列表缓存
func (c *BusinessCache) GetUserPermissions(ctx context.Context, userID uint64) ([]string, error) {
	key := fmt.Sprintf("%s%d", CacheKeyUserPermissions, userID)
	var permissions []string
	err := c.Get(ctx, key, &permissions)
	if err == ErrCacheMiss {
		return nil, ErrCacheMiss
	}
	return permissions, err
}

func (c *BusinessCache) SetUserPermissions(ctx context.Context, userID uint64, permissions []string) error {
	key := fmt.Sprintf("%s%d", CacheKeyUserPermissions, userID)
	return c.Set(ctx, key, permissions, CacheExpireUserPermissions)
}

func (c *BusinessCache) DeleteUserPermissions(ctx context.Context, userID uint64) error {
	key := fmt.Sprintf("%s%d", CacheKeyUserPermissions, userID)
	return c.Delete(ctx, key)
}

// 用户菜单树缓存
func (c *BusinessCache) GetUserMenuTree(ctx context.Context, userID uint64, v interface{}) error {
	key := fmt.Sprintf("%s%d", CacheKeyUserMenuTree, userID)
	return c.Get(ctx, key, v)
}

func (c *BusinessCache) SetUserMenuTree(ctx context.Context, userID uint64, v interface{}) error {
	key := fmt.Sprintf("%s%d", CacheKeyUserMenuTree, userID)
	return c.Set(ctx, key, v, CacheExpireUserMenuTree)
}

func (c *BusinessCache) DeleteUserMenuTree(ctx context.Context, userID uint64) error {
	key := fmt.Sprintf("%s%d", CacheKeyUserMenuTree, userID)
	return c.Delete(ctx, key)
}

// 完整菜单树缓存
func (c *BusinessCache) GetMenuTree(ctx context.Context, v interface{}) error {
	return c.Get(ctx, CacheKeyMenuTree, v)
}

func (c *BusinessCache) SetMenuTree(ctx context.Context, v interface{}) error {
	return c.Set(ctx, CacheKeyMenuTree, v, CacheExpireMenuTree)
}

func (c *BusinessCache) DeleteMenuTree(ctx context.Context) error {
	return c.Delete(ctx, CacheKeyMenuTree)
}

// 字典项列表缓存（按 code）
func (c *BusinessCache) GetDictItems(ctx context.Context, code string, v interface{}) error {
	key := fmt.Sprintf("%s%s", CacheKeyDictItems, code)
	return c.Get(ctx, key, v)
}

func (c *BusinessCache) SetDictItems(ctx context.Context, code string, v interface{}) error {
	key := fmt.Sprintf("%s%s", CacheKeyDictItems, code)
	return c.Set(ctx, key, v, CacheExpireDictItems)
}

func (c *BusinessCache) DeleteDictItems(ctx context.Context, code string) error {
	key := fmt.Sprintf("%s%s", CacheKeyDictItems, code)
	return c.Delete(ctx, key)
}

// 字典项列表缓存（按 type_id）
func (c *BusinessCache) GetDictItemsByType(ctx context.Context, typeID uint64, v interface{}) error {
	key := fmt.Sprintf("%s%d", CacheKeyDictItemsByType, typeID)
	return c.Get(ctx, key, v)
}

func (c *BusinessCache) SetDictItemsByType(ctx context.Context, typeID uint64, v interface{}) error {
	key := fmt.Sprintf("%s%d", CacheKeyDictItemsByType, typeID)
	return c.Set(ctx, key, v, CacheExpireDictItems)
}

func (c *BusinessCache) DeleteDictItemsByType(ctx context.Context, typeID uint64) error {
	key := fmt.Sprintf("%s%d", CacheKeyDictItemsByType, typeID)
	return c.Delete(ctx, key)
}

// 配置缓存
func (c *BusinessCache) GetConfigKey(ctx context.Context, key string, v interface{}) error {
	cacheKey := fmt.Sprintf("%s%s", CacheKeyConfigKey, key)
	return c.Get(ctx, cacheKey, v)
}

func (c *BusinessCache) SetConfigKey(ctx context.Context, key string, v interface{}) error {
	cacheKey := fmt.Sprintf("%s%s", CacheKeyConfigKey, key)
	return c.Set(ctx, cacheKey, v, CacheExpireConfigKey)
}

func (c *BusinessCache) DeleteConfigKey(ctx context.Context, key string) error {
	cacheKey := fmt.Sprintf("%s%s", CacheKeyConfigKey, key)
	return c.Delete(ctx, cacheKey)
}
