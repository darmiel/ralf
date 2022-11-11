package parse

import (
	"github.com/ralf-life/engine/model"
	"gopkg.in/yaml.v3"
)

type parseCondition struct {
	If   string
	Then []yaml.Node
	Else []yaml.Node
}

func parseConditionFromYaml(node *yaml.Node) (model.Flow, error) {
	var (
		err       error
		parseCond *parseCondition
	)
	if err = node.Decode(&parseCond); err != nil {
		panic(err)
	}
	// fill then's
	var thenFlows []model.Flow
	for _, condThen := range parseCond.Then {
		flow, err := ParseFlow(&condThen)
		if err != nil {
			return nil, err
		}
		thenFlows = append(thenFlows, flow)
	}
	// fill else's
	var elseFlows []model.Flow
	for _, condElse := range parseCond.Else {
		flow, err := ParseFlow(&condElse)
		if err != nil {
			return nil, err
		}
		elseFlows = append(elseFlows, flow)
	}
	return &model.ConditionFlow{
		Condition: parseCond.If,
		Then:      thenFlows,
		Else:      elseFlows,
	}, nil
}
