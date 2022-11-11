package model

import (
	"time"
)

type Profile struct {
	Name          string        `yaml:"name"`
	CacheDuration time.Duration `yaml:"cache-duration"`
	Flows         []Flow        `yaml:"-"`
}

type ConditionFlow struct {
	Condition string `yaml:"if"`
	Then      []Flow `yaml:"-"`
	Else      []Flow `yaml:"-"`
}

type ActionFlow struct {
	FlowIdentifier string                 `yaml:"do"`
	With           map[string]interface{} `yaml:"with"`
}
