package model

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v3"
)

// Flow can represent an action, a condition or a return statement
type Flow interface {
	KeyIdentifier() string
}

type Flows []Flow

///

// ActionFlow represents an action to run more flows
type ActionFlow struct {
	FlowIdentifier string                 `yaml:"do" json:"do" bson:"do"`
	With           map[string]interface{} `yaml:"with" json:"with" bson:"with"`
}

func (a *ActionFlow) KeyIdentifier() string {
	return "do"
}

///

type Conditions []string

func (c *Conditions) MarshalYAML() (interface{}, error) {
	if len(*c) == 1 {
		return (*c)[0], nil
	}
	return c, nil
}

func (c *Conditions) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!!str" {
		var single string
		if err := value.Decode(&single); err != nil {
			return err
		}
		*c = []string{single}
	} else {
		var multi []string
		if err := value.Decode(&multi); err != nil {
			return err
		}
		*c = multi
	}
	return nil
}

func (c *Conditions) MarshalJSON() ([]byte, error) {
	if len(*c) == 1 {
		return json.Marshal((*c)[0])
	}
	return json.Marshal([]string(*c))
}

func (c *Conditions) UnmarshalJSON(val []byte) error {
	if bytes.HasPrefix(val, []byte("\"")) {
		var single string
		if err := json.Unmarshal(val, &single); err != nil {
			return err
		}
		*c = []string{single}
	} else {
		var multi []string
		if err := json.Unmarshal(val, &multi); err != nil {
			return err
		}
		*c = multi
	}
	return nil
}

// ConditionFlow represents a condition
type ConditionFlow struct {
	Condition Conditions `yaml:"if" json:"if" bson:"if"`
	Operator  string     `yaml:"op" json:"op" bson:"op"`
	Then      Flows      `yaml:"then" json:"then" bson:"then"`
	Else      Flows      `yaml:"else" json:"else" bson:"else"`
}

func (c *ConditionFlow) KeyIdentifier() string {
	return "if"
}

///

// ReturnFlow stops the current execution immediately
type ReturnFlow struct {
	Return bool `yaml:"return" json:"return" bson:"return"`
}

func (r *ReturnFlow) KeyIdentifier() string {
	return "return"
}

///

// DebugFlow prints a message in the console
type DebugFlow struct {
	Debug interface{} `yaml:"debug" json:"debug" bson:"debug"`
}

func (d *DebugFlow) KeyIdentifier() string {
	return "debug"
}
