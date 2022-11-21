package engine

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	ics "github.com/darmiel/golang-ical"
	"github.com/ralf-life/engine/actions"
	"github.com/ralf-life/engine/model"
	"strings"
)

type ContextFlow struct {
	*model.Profile
	Context map[string]interface{}
	Debugs  []interface{}
}

var dummyContextEnv = &ContextEnv{}

func (c *ContextFlow) RunSingleFlow(event *ics.VEvent, flow model.Flow) (ExecutionMessage, error) {
	switch f := flow.(type) {

	// ReturnFlow:
	// Exit loop
	case *model.ReturnFlow:
		return &ExitMessage{}, nil

	// DebugFlow:
	// Print message to console
	case *model.DebugFlow:
		if str, ok := f.Debug.(string); ok {
			// ${Date} is ${Date.IsAfter("9:00")}
			if strings.HasPrefix(str, "$ ") {
				ex, err := expr.Compile(str[2:], expr.Env(dummyContextEnv))
				if err != nil {
					return nil, err
				}
				env, err := c.CreateEnv(event)
				if err != nil {
					return nil, err
				}
				res, err := expr.Run(ex, env)
				if err != nil {
					return nil, err
				}
				return &DebugMessage{Message: res}, nil
			}
		}
		return &DebugMessage{f.Debug}, nil

	// ConditionFlow:
	// Check condition and execute child flows
	case *model.ConditionFlow:
		ex, err := expr.Compile(f.Condition, expr.Env(dummyContextEnv), expr.AsBool())
		if err != nil {
			return nil, err
		}
		env, err := c.CreateEnv(event)
		if err != nil {
			return nil, err
		}
		res, err := expr.Run(ex, env)
		if err != nil {
			return nil, err
		}
		// queue flow children
		if res.(bool) {
			return &QueueMessage{f.Then}, nil
		} else {
			return &QueueMessage{f.Else}, nil
		}

	case *model.ActionFlow:
		// find action
		act := actions.Find(f.FlowIdentifier)
		if act == nil {
			return nil, errors.New("invalid flow identifier: " + f.FlowIdentifier)
		}
		msg, err := act.Execute(event, f.With)
		if err != nil {
			return nil, err
		}
		// all good if no message was returned
		if msg == nil {
			return nil, nil
		}
		switch msg.(type) {
		case actions.FilterInMessage:
			return &FilterMessage{actions.DummyFilterInMessage}, nil
		case actions.FilterOutMessage:
			return &FilterMessage{actions.DummyFilterOutMessage}, nil
		default:
			panic("invalid type for model.ActionFlow->Execute->msg")
		}
	}
	return nil, nil
}

var ErrExited = errors.New("flows exited because of a return statement")

func (c *ContextFlow) runCycleFlows(fact *actions.ActionMessage, event *ics.VEvent, flows model.Flows) error {
	for _, flow := range flows {
		msg, err := c.RunSingleFlow(event, flow)
		// oh no, we always exit on errors
		if err != nil {
			return err
		}
		// if msg is null, all good and continue loop
		if msg == nil {
			continue
		}
		switch t := msg.(type) {
		// exit loop
		case *ExitMessage:
			return ErrExited
		case *QueueMessage:
			if err = c.runCycleFlows(fact, event, t.Flows); err != nil {
				// child process exited (or failed)
				// -> also exit all parent flows
				return err
			}
		case *FilterMessage:
			*fact = t.Action
		case *DebugMessage:
			fmt.Println("[DEBUG]", t.Message)
			c.Debugs = append(c.Debugs, t.Message)
		}
	}
	return nil
}

func (c *ContextFlow) RunAllFlows(event *ics.VEvent, flows model.Flows) (actions.ActionMessage, error) {
	var fact actions.ActionMessage = actions.DummyFilterInMessage
	err := c.runCycleFlows(&fact, event, flows)
	return fact, err
}
