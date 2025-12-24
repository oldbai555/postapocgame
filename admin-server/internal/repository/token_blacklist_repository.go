package repository

import (
	"context"
	"time"

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
	res, err := r.repo.Redis.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.Wrap(err, "redis exists token blacklist")
	}
	return res > 0, nil
}

func (r *tokenBlacklistRepository) Blacklist(ctx context.Context, token string, ttl time.Duration) error {
	if token == "" {
		return nil
	}
	key := blacklistKey(token)
	if err := r.repo.Redis.SetEx(ctx, key, "1", ttl).Err(); err != nil {
		return errors.Wrap(err, "redis setex token blacklist")
	}
	return nil
}

func blacklistKey(token string) string {
	return "jwt:blacklist:" + token
}
