package model

import (
	"gopkg.in/yaml.v3"
)

type Flows []Flow

func (f *Flows) UnmarshalJSON(data []byte) error {
	return nil
}

func (f *Flows) UnmarshalYAML(value *yaml.Node) error {
	return nil
}

// Flow can represent an action, a condition or a return statement
type Flow interface {
	KeyIdentifier() string
}

// ConditionFlow represents a condition
type ConditionFlow struct {
	Condition string `yaml:"if"`
	Then      []Flow `yaml:"then"`
	Else      []Flow `yaml:"else"`
}

func (c *ConditionFlow) KeyIdentifier() string {
	return "if"
}

// ActionFlow represents an action to run more flows
type ActionFlow struct {
	FlowIdentifier string                 `yaml:"do"`
	With           map[string]interface{} `yaml:"with"`
}

func (a *ActionFlow) KeyIdentifier() string {
	return "do"
}

// ReturnFlow stops the current execution immediately
type ReturnFlow struct {
}

func (r *ReturnFlow) KeyIdentifier() string {
	return "return"
}

// DebugFlow prints a message in the console
type DebugFlow struct {
	Message interface{}
}

func (d *DebugFlow) KeyIdentifier() string {
	return "debug"
}

/// Predefined flows

var (
	Return = &ReturnFlow{}
)
