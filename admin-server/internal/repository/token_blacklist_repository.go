package repository

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/consts"

	"github.com/pkg/errors"
)

type TokenBlacklistRepository interface {
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	Blacklist(ctx context.Context, token string, ttl time.Duration) error
}

type tokenBlacklistRepository struct {
	repo *Repository
}

func NewTokenBlacklistRepository(repo *Repository) TokenBlacklistRepository {
	return &tokenBlacklistRepository{repo: repo}
}

func (r *tokenBlacklistRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}
	key := blacklistKey(token)
	// go-zero Redis Exists 返回 bool
	exists, err := r.repo.Redis.Exists(key)
	if err != nil {
		return false, errors.Wrap(err, "redis exists token blacklist")
	}
	return exists, nil
}

func (r *tokenBlacklistRepository) Blacklist(ctx context.Context, token string, ttl time.Duration) error {
	if token == "" {
		return nil
	}
	key := blacklistKey(token)
	// go-zero Redis Setex 方法名是小写 x，参数：key, value, seconds
	if err := r.repo.Redis.Setex(key, "1", int(ttl.Seconds())); err != nil {
		return errors.Wrap(err, "redis setex token blacklist")
	}
	return nil
}

func blacklistKey(token string) string {
	return consts.RedisJWTBlacklistPrefix + token
}
