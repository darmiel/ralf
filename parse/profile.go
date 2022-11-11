package parse

import (
	"github.com/ralf-life/engine/model"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type parseProfile struct {
	Name          string        `yaml:"name"`
	CacheDuration time.Duration `yaml:"cache-duration"`
	Flows         []yaml.Node   `yaml:"flows"`
}

func ParseProfile(file string) (*model.Profile, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	var p parseProfile
	if err = dec.Decode(&p); err != nil {
		panic(err)
	}

	var flows []model.Flow
	for _, flow := range p.Flows {
		flow, err := ParseFlow(&flow)
		if err != nil {
			panic(err)
		}
		flows = append(flows, flow)
	}

	return &model.Profile{
		Name:          p.Name,
		CacheDuration: p.CacheDuration,
		Flows:         flows,
	}, nil
}
