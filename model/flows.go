package model

// Flow can represent an action, a condition or a return statement
type Flow interface {
	KeyIdentifier() string
}

// ActionFlow represents an action to run more flows
type ActionFlow struct {
	FlowIdentifier string                 `yaml:"do"`
	With           map[string]interface{} `yaml:"with"`
}

func (a *ActionFlow) KeyIdentifier() string {
	return "do"
}

// ConditionFlow represents a condition
type ConditionFlow struct {
	Condition string `yaml:"if"`
	Then      Flows  `yaml:"then"`
	Else      Flows  `yaml:"else"`
}

func (c *ConditionFlow) KeyIdentifier() string {
	return "if"
}

// ReturnFlow stops the current execution immediately
type ReturnFlow struct {
}

func (r *ReturnFlow) KeyIdentifier() string {
	return "return"
}

// DebugFlow prints a message in the console
type DebugFlow struct {
	Debug interface{}
}

func (d *DebugFlow) KeyIdentifier() string {
	return "debug"
}

/// Predefined flows

var (
	Return = &ReturnFlow{}
)
