package model

import (
	"encoding/json"
	"errors"
	"fmt"
	ics "github.com/darmiel/golang-ical"
	"github.com/darmiel/ralf/pkg/sources"
	htmlsource "github.com/darmiel/ralf/pkg/sources/html"
	httpsource "github.com/darmiel/ralf/pkg/sources/http"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v3"
)

var (
	ErrUnknownSourceType = errors.New("unknown sources type")
	ErrInvalidLength     = errors.New("only one sources is allowed")
)

var sourceTypes = []Source{
	&httpsource.Source{},
	&htmlsource.Source{},
}

// Source is an interface for different types of sources, such as HTTP or HTML sources.
type Source interface {
	KeyIdentifier() string
	Validate() error
	CacheKey() (string, error)
	Run() (*ics.Calendar, error)
}

type SomeSource []Source

func newLegacySource(url string) Source {
	return &httpsource.Source{
		Type: sources.Type{
			Type: "http",
		},
		URL: url,
	}
}

func (s *SomeSource) UnmarshalJSON(data []byte) error {
	var url string

	// Attempt to unmarshal data as a simple URL (legacy support).
	if err := json.Unmarshal(data, &url); err == nil {
		*s = append(*s, newLegacySource(url))
		return nil
	}

	// Attempt to unmarshal data as a SourceType to determine the specific type of sources.
	var src sources.Type
	if err := json.Unmarshal(data, &src); err != nil {
		return err
	}

	for _, t := range sourceTypes {
		if t.KeyIdentifier() != src.Type {
			continue
		}
		c := cloneSource(t)
		if err := json.Unmarshal(data, c); err != nil {
			return err
		}
		*s = append(*s, c)
		return nil
	}

	return ErrUnknownSourceType
}

func (s *SomeSource) MarshalJSON() ([]byte, error) {
	if len(*s) != 1 {
		return nil, ErrInvalidLength
	}
	return json.Marshal((*s)[0])
}

func (s *SomeSource) UnmarshalYAML(value *yaml.Node) error {
	var url string

	// Attempt to unmarshal data as a simple URL (legacy support).
	if err := value.Decode(&url); err == nil {
		*s = append(*s, newLegacySource(url))
		return nil
	}

	// Attempt to unmarshal data as a SourceType to determine the specific type of sources.
	var src sources.Type
	if err := value.Decode(&src); err != nil {
		return err
	}

	for _, t := range sourceTypes {
		if t.KeyIdentifier() != src.Type {
			continue
		}
		c := cloneSource(t)
		if err := value.Decode(c); err != nil {
			return err
		}
		*s = append(*s, c)
		return nil
	}

	return fmt.Errorf("%w: %s", ErrUnknownSourceType, src.Type)
}

func (s *SomeSource) MarshalYAML() (interface{}, error) {
	if len(*s) != 1 {
		return nil, ErrInvalidLength
	}
	return yaml.Marshal((*s)[0])
}

func (s *SomeSource) UnmarshalBSON(data []byte) error {
	var url string

	// Attempt to unmarshal data as a simple URL (legacy support).
	if err := bson.Unmarshal(data, &url); err == nil {
		*s = append(*s, newLegacySource(url))
		return nil
	}

	// Attempt to unmarshal data as a SourceType to determine the specific type of sources.
	var src sources.Type
	if err := bson.Unmarshal(data, &src); err != nil {
		return err
	}

	for _, t := range sourceTypes {
		if t.KeyIdentifier() != src.Type {
			continue
		}
		c := cloneSource(t)
		if err := bson.Unmarshal(data, c); err != nil {
			return err
		}
		*s = append(*s, c)
		return nil
	}

	return fmt.Errorf("%w: %s", ErrUnknownSourceType, src.Type)
}

func (s *SomeSource) MarshalBSON() ([]byte, error) {
	if len(*s) != 1 {
		return nil, ErrInvalidLength
	}
	return bson.Marshal((*s)[0])
}

// cloneSource creates a clone of a Source interface to prevent data races and unintended modifications.
func cloneSource(src Source) Source {
	switch v := src.(type) {
	case *httpsource.Source:
		clone := *v
		return &clone
	case *htmlsource.Source:
		clone := *v
		return &clone
	default:
		panic(fmt.Sprintf("unknown source type: %T", src))
	}
}
