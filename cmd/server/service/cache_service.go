package service

import "context"

// CacheService defines the interface for cache services.
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, duration int64) error
}
