package model

// Profile represents a filter profile
type Profile struct {
	Name          string     `yaml:"name" json:"name"`
	Source        SomeSource `yaml:"source" json:"source"`
	CacheDuration Duration   `yaml:"cache-duration" json:"cache-duration"`
	Flows         Flows      `yaml:"flows" json:"flows"`
}
