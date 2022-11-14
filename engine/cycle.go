package engine

import (
	"errors"
	"fmt"
	"github.com/ralf-life/engine/model"
)

type ContextFlow struct {
	*model.Profile
	Context map[string]interface{}
}

func (c *ContextFlow) Run(flow model.Flow) (ExecutionMessage, error) {
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
	}
	return nil, nil
}

var ErrExited = errors.New("flows exited because of a return statement")

func (c *ContextFlow) RunCycleFlows(flows model.Flows) error {
	for _, flow := range flows {
		msg, err := c.Run(flow)
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
			if err = c.RunCycleFlows(t.Flows); err != nil {
				// child process exited (or failed)
				// -> also exit all parent flows
				return err
			}
		}
	}
	return nil
}
