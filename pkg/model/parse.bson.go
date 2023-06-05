package model

import (
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrBsonKeyNotFound = errors.New("unknown type converter for BSON type")

type bsonConverterFun func([]byte) (Flow, error)

func convertToFlow[T Flow](data []byte) (Flow, error) {
	var t T
	err := json.Unmarshal(data, &t)
	return t, err
}

func convertFun[T Flow]() func([]byte) (Flow, error) {
	return func(data []byte) (Flow, error) {
		return convertToFlow[T](data)
	}
}

var bsonKeys = map[string]bsonConverterFun{
	// condition flow
	"if":     convertFun[*ConditionFlow](),
	"do":     convertFun[*ActionFlow](),
	"debug":  convertFun[*DebugFlow](),
	"return": convertFun[*ReturnFlow](),
}

func convertRawBSONToFlow(v bson.RawValue) (Flow, error) {
	var m map[string]interface{}
	if err := v.Unmarshal(&m); err != nil {
		return nil, err
	}
	for k := range m {
		if fun, ok := bsonKeys[k]; ok {
			data, err := json.Marshal(m)
			if err != nil {
				return nil, err
			}
			res, err := fun(data)
			if err != nil {
				return nil, err
			}
			return res, nil
		}
	}
	return nil, ErrBsonKeyNotFound
}

// UnmarshalBSON is a little big illegal.
// It
//  1. unmarshalls BSON to a "raw document"
//  2. unmarshalls every value from the raw document to a map[string]any
//  3. marshals this list to JSON
//  4. unmarshalls the JSON to the desired type
func (f *Flows) UnmarshalBSON(b []byte) error {
	var raw bson.Raw
	if err := bson.Unmarshal(b, &raw); err != nil {
		return err
	}
	values, err := raw.Values()
	if err != nil {
		return err
	}
	var res Flows
	for _, v := range values {
		f, err := convertRawBSONToFlow(v)
		if err != nil {
			return err
		}
		res = append(res, f)
	}
	*f = res
	return nil
}
