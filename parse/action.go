package parse

import (
	"github.com/ralf-life/engine/model"
	"gopkg.in/yaml.v3"
)

func parseActionFromYaml(node *yaml.Node) (model.Flow, error) {
	var action *model.ActionFlow
	if err := node.Decode(&action); err != nil {
		return nil, err
	}
	return action, nil
}
