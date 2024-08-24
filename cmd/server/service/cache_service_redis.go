package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// make sure RedisCacheService implements CacheService
var _ CacheService = (*RedisCacheService)(nil)

// RedisCacheService implements CacheService using Redis.
type RedisCacheService struct {
	client *redis.Client
}

// NewRedisCacheService creates a new RedisCacheService.
func NewRedisCacheService(addr, password string, db int) *RedisCacheService {
	return &RedisCacheService{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

// Get retrieves a value from Redis by key.
func (r *RedisCacheService) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set sets a value in Redis with the specified expiration duration.
func (r *RedisCacheService) Set(ctx context.Context, key string, value string, duration int64) error {
	return r.client.Set(ctx, key, value, time.Duration(duration)*time.Second).Err()
}
