package model

// Profile represents a filter profile
type Profile struct {
	Name          string   `yaml:"name" json:"name"`
	CacheDuration Duration `yaml:"cache-duration" json:"cache-duration"`
	Flows         Flows    `yaml:"flows" json:"flows"`
}
