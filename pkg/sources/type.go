package sources

// Type represents a type of sources with a specific key.
// It is only used as a helper to unmarshal the sources type.
type Type struct {
	Type string `json:"type" yaml:"type"`
}
