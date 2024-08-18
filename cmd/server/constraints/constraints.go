package constraints

import (
	"github.com/darmiel/ralf/pkg/model"
	"time"
)

// ClampCacheDuration clamps the cache duration to a minimum of 2 minutes and a maximum of 24 hours.
func ClampCacheDuration(cacheDuration time.Duration) time.Duration {
	if cacheDuration.Minutes() < 2.0 {
		return 2 * time.Minute
	}
	if cacheDuration.Hours() > 24.0 {
		return 24 * time.Hour
	}
	return cacheDuration
}

// ClampCacheModelDuration clamps the cache duration to a minimum of 2 minutes and a maximum of 24 hours.
func ClampCacheModelDuration(cacheDuration model.Duration) model.Duration {
	return model.Duration(ClampCacheDuration(time.Duration(cacheDuration)))
}

// ClampHistoryLimit clamps the history limit to a minimum of 1 and a maximum of 1000.
func ClampHistoryLimit(limit int) int {
	if limit < 1 {
		return 1
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}
