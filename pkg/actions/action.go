package actions

import (
	"errors"
	"fmt"
	ics "github.com/darmiel/golang-ical"
)

var Actions = []Action{
	new(FilterInAction),
	new(FilterOutAction),
	new(RegexReplaceAction),
	new(ClearAttendeesAction),
	new(AddAttendeeAction),
	new(ClearAlarmsAction),
	new(AddAlarmAction),
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
	FilterOutActionMessage struct{}
	FilterInActionMessage  struct{}
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

func has(with map[string]interface{}, key string) (ok bool) {
	_, ok = with[key]
	return
}

func strArray(with map[string]interface{}, key string, def []interface{}) ([]string, error) {
	in, err := optional[[]interface{}](with, key, def)
	if err != nil {
		return nil, err
	}
	var res = make([]string, len(in))
	for i, v := range in {
		str, ok := v.(string)
		if !ok {
			return nil, errors.New(fmt.Sprintf("'%v' is not a string", v))
		}
		res[i] = str
	}
	return res, nil
}
