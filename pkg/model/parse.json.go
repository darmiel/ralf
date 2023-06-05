package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func jsonConverterFun[T Flow]() func(message *json.RawMessage) (Flow, error) {
	return func(msg *json.RawMessage) (Flow, error) {
		var cond T
		err := json.Unmarshal(*msg, &cond)
		return cond, err
	}
}

var jsonKeys = map[string]func(msg *json.RawMessage) (Flow, error){
	// condition flow
	"if":     jsonConverterFun[*ConditionFlow](),
	"do":     jsonConverterFun[*ActionFlow](),
	"debug":  jsonConverterFun[*DebugFlow](),
	"return": jsonConverterFun[*ReturnFlow](),
}

///

// Duration is required since we cannot simply unmarshal `time.Duration`
// https://stackoverflow.com/a/54571600/10564458
type Duration time.Duration

// MarshalJSON marshals a time.Duration into JSON
func (d *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(*d).String())
}

// UnmarshalJSON un-marshals a time.Duration from JSON
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

///

func convertRawJSONToFlow(msg *json.RawMessage) (Flow, error) {
	var kv map[string]interface{}
	if err := json.Unmarshal(*msg, &kv); err != nil {
		return nil, err
	}
	for k, fun := range jsonKeys {
		if _, ok := kv[k]; ok {
			return fun(msg)
		}
	}
	return nil, errors.New("unknown map: " + fmt.Sprintf("%+v", kv))
}

func (f *Flows) UnmarshalJSON(data []byte) error {
	// fmt.Println("called unmarshal json with data", string(data))
	var arr []*json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	for _, a := range arr {
		flow, err := convertRawJSONToFlow(a)
		if err != nil {
			return err
		}
		*f = append(*f, flow)
	}
	return nil
}

func ParseProfileFromJSON(data []byte) (*Profile, error) {
	var prof Profile
	if err := json.Unmarshal(data, &prof); err != nil {
		return nil, err
	}
	return &prof, nil
}
