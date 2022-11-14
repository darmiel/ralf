package parse

import (
	"errors"
	"fmt"
	"github.com/ralf-life/engine/model"
	"gopkg.in/yaml.v3"
)

func FlowFromYaml(node *yaml.Node) (model.Flow, error) {
	// simple text flow
	if v := node.Value; v != "" {
		switch v {
		case "return":
			return model.Return, nil
		}
	}

	var err error

	// mapped flow
	if node.Tag == "!!map" {
		var kv map[string]interface{}
		if err = node.Decode(&kv); err != nil {
			return nil, err
		}
		// condition
		if _, ok := kv["if"]; ok {
			return parseConditionFromYaml(node)
		} else if _, ok = kv["do"]; ok {
			return parseActionFromYaml(node)
		} else if _, ok = kv["debug"]; ok {
			return &model.DebugFlow{Message: kv["debug"]}, nil
		} else {
			panic("unknown map: " + fmt.Sprintf("%+v", kv))
		}
	}

	return nil, errors.New("invalid tag: " + node.Tag)
}
