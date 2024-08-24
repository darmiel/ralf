package service

import (
	"context"
	"github.com/patrickmn/go-cache"
	"time"
)

// make sure LocalCacheService implements CacheService
var _ CacheService = (*LocalCacheService)(nil)

// LocalCacheService implements CacheService using an in-memory cache.
type LocalCacheService struct {
	cache *cache.Cache
}

// NewLocalCacheService creates a new LocalCacheService.
func NewLocalCacheService(defaultExpiration, cleanupInterval time.Duration) *LocalCacheService {
	return &LocalCacheService{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves a value from the cache by key.
func (l *LocalCacheService) Get(_ context.Context, key string) (string, error) {
	value, found := l.cache.Get(key)
	if !found {
		return "", nil
	}
	return value.(string), nil
}

// Set sets a value in the cache with the specified expiration duration.
func (l *LocalCacheService) Set(_ context.Context, key string, value string, duration int64) error {
	l.cache.Set(key, value, time.Duration(duration)*time.Second)
	return nil
}
