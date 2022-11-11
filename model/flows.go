package model

type Flow interface{}

type ReturnFlow struct{}

var Return = &ReturnFlow{}

type DebugFlow struct {
	Message interface{}
}
