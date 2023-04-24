package actions

import (
	"errors"
	"fmt"
	ics "github.com/darmiel/golang-ical"
)

var Actions = []Action{
	&FilterInAction{},
	&FilterOutAction{},
	&RegexReplaceAction{},
}

func Find(identifier string) Action {
	for _, act := range Actions {
		if act.Identifier() == identifier {
			return act
		}
	}
	return nil
}

type Action interface {
	Identifier() string
	Execute(event *ics.VEvent, with map[string]interface{}, verbose bool) (ActionMessage, error)
}

type ActionMessage interface {
}

type (
	FilterOutMessage byte
	FilterInMessage  byte
)

func required[T any](with map[string]interface{}, key string) (T, error) {
	ifa, ok := with[key]
	if !ok {
		var n T
		return n, errors.New(fmt.Sprintf("'%s' required.", key))
	}
	val, ok := ifa.(T)
	if !ok {
		var n T
		return n, errors.New(fmt.Sprintf("'%s' has an invalid type", key))
	}
	return val, nil
}

func optional[T any](with map[string]interface{}, key string, def T) (T, error) {
	ifa, ok := with[key]
	if !ok {
		return def, nil
	}
	val, ok := ifa.(T)
	if !ok {
		var n T
		return n, errors.New(fmt.Sprintf("'%s' must be a %T", key, def))
	}
	return val, nil
}
