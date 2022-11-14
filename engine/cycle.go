package engine

import (
	"errors"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/ralf-life/engine/actions"
	"github.com/ralf-life/engine/model"
)

type ContextFlow struct {
	*model.Profile
	Context map[string]interface{}
}

func (c *ContextFlow) RunSingleFlow(event *ics.VEvent, flow model.Flow) (ExecutionMessage, error) {
	switch f := flow.(type) {

	// ReturnFlow:
	// Exit loop
	case *model.ReturnFlow:
		return &ExitMessage{}, nil

	// DebugFlow:
	// Print message to console
	case *model.DebugFlow:
		fmt.Println("[DEBUG]", f.Debug)
		return nil, nil

	// ConditionFlow:
	// Check condition and execute child flows
	case *model.ConditionFlow:
		var out bool
		// TODO: execute condition
		out = f.Condition == "true"
		// queue flow children
		if out {
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
			fmt.Printf("[FILTER] Changed state to %T\n", t.Action)
		}
	}
	return nil
}

func (c *ContextFlow) RunAllFlows(event *ics.VEvent, flows model.Flows) (actions.ActionMessage, error) {
	var fact actions.ActionMessage = actions.DummyFilterInMessage
	err := c.runCycleFlows(&fact, event, flows)
	return fact, err
}
