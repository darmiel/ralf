package model

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"time"
)

func ParseProfileFromYAML(reader io.Reader) (*Profile, error) {
	dec := yaml.NewDecoder(reader)
	dec.KnownFields(true)

	var p Profile
	err := dec.Decode(&p)
	return &p, err
}

func (f *Flows) UnmarshalYAML(value *yaml.Node) error {
	for _, child := range value.Content {
		flow, err := yamlParseNode(child)
		if err != nil {
			return err
		}
		*f = append(*f, flow)
	}
	return nil
}

var yamlKeys = map[string]func(node *yaml.Node) (Flow, error){
	// condition flow
	"if": func(node *yaml.Node) (Flow, error) {
		var cond *ConditionFlow
		err := node.Decode(&cond)
		return cond, err
	},
	"do": func(node *yaml.Node) (Flow, error) {
		var act *ActionFlow
		err := node.Decode(&act)
		return act, err
	},
	"debug": func(node *yaml.Node) (Flow, error) {
		var dbg *DebugFlow
		err := node.Decode(&dbg)
		return dbg, err
	},
	"return": func(node *yaml.Node) (Flow, error) {
		var rtf *ReturnFlow
		err := node.Decode(&rtf)
		return rtf, err
	},
}

var yamlTags = map[string]func(node *yaml.Node) (Flow, error){
	"!!map": func(node *yaml.Node) (Flow, error) {
		var (
			err error
			kv  map[string]interface{}
		)
		if err = node.Decode(&kv); err != nil {
			return nil, err
		}
		for k, fun := range yamlKeys {
			if _, ok := kv[k]; ok {
				return fun(node)
			}
		}
		return nil, errors.New("unknown map: " + fmt.Sprintf("%+v", kv))
	},
}

func yamlParseNode(node *yaml.Node) (Flow, error) {
	fun, ok := yamlTags[node.Tag]
	if !ok {
		return nil, errors.New("invalid tag: " + node.Tag)
	}
	// try to parse from func
	return fun(node)
}

func (d *Duration) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(time.Duration(*d).String())
}

func (d *Duration) UnmarshalYAML(b *yaml.Node) error {
	var v interface{}
	if err := b.Decode(&v); err != nil {
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
