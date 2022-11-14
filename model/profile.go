package model

import (
	"time"
)

// Profile represents a filter profile
type Profile struct {
	Name          string        `yaml:"name"`
	CacheDuration time.Duration `yaml:"cache-duration"`
	Flows         []Flow        `yaml:"flows"`
}
