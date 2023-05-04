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
	Context     map[string]interface{}
	EnableDebug bool
	Verbose     bool
	Debugs      []interface{}
}

var ErrExited = errors.New("flows exited because of a return statement")

func runSingleDebugFlow(f *model.DebugFlow, e *ics.VEvent) (ExecutionMessage, error) {
	if str, ok := f.Debug.(string); ok {
		// evaluated debug messages can start with "$"
		if strings.HasPrefix(str, "$ ") {
			ex, err := expr.Compile(str[2:], expr.Env(new(ExprEnvironment)))
			if err != nil {
				return nil, err
			}
			env, err := CreateExprEnvironmentFromEvent(e)
			if err != nil {
				return nil, err
			}
			res, err := expr.Run(ex, env)
			if err != nil {
				return nil, err
			}
			return &DebugExecutionMessage{Message: res}, nil
		}
	}
	return &DebugExecutionMessage{f.Debug}, nil
}

func runSingleConditionFlow(f *model.ConditionFlow, e *ics.VEvent) (ExecutionMessage, error) {
	env, err := CreateExprEnvironmentFromEvent(e)
	if err != nil {
		return nil, fmt.Errorf("create expr env err: %v", err)
	}

	result := false
	isAnd := strings.ToUpper(f.Operator) == "AND"

	for _, cond := range f.Condition {
		ex, err := expr.Compile(cond, expr.Env(new(ExprEnvironment)), expr.AsBool())
		if err != nil {
			return nil, fmt.Errorf("expr compile err: %v", err)
		}
		res, err := expr.Run(ex, env)
		if err != nil {
			return nil, fmt.Errorf("expr run err: %v", err)
		}
		if !res.(bool) && isAnd {
			result = false
			break
		} else if res.(bool) {
			result = true
		}
	}
	if !result {
		return &QueueFlowsExecutionMessage{f.Else}, nil
	}
	return &QueueFlowsExecutionMessage{f.Then}, nil

}

func runSingleActionFlow(f *model.ActionFlow, e *ics.VEvent, verbose bool) (ExecutionMessage, error) {
	// find action
	act := actions.Find(f.FlowIdentifier)
	if act == nil {
		return nil, errors.New("invalid flow identifier: " + f.FlowIdentifier)
	}
	msg, err := act.Execute(e, f.With, verbose)
	if err != nil {
		return nil, fmt.Errorf("flow execute err: %v", err)
	}
	if msg != nil {
		switch msg.(type) {
		case *actions.FilterInActionMessage, *actions.FilterOutActionMessage:
			return &FilterResultExecutionMessage{Action: msg}, nil
		default:
			panic("invalid type for model.ActionFlow->Execute->msg")
		}
	}
	// all good if no message was returned
	return nil, nil
}

func RunSingleFlow(event *ics.VEvent, flow model.Flow, verbose, enableDebugFlow bool) (ExecutionMessage, error) {
	switch f := flow.(type) {

	// ReturnFlow:
	// Exit loop
	case *model.ReturnFlow:
		return new(ExitFlowsExecutionMessage), nil

	// DebugFlow:
	// Print message to console
	case *model.DebugFlow:
		if !enableDebugFlow {
			return nil, nil
		}
		return runSingleDebugFlow(f, event)

	// ConditionFlow:
	// Check condition and execute child flows
	case *model.ConditionFlow:
		return runSingleConditionFlow(f, event)

	// ActionFlow
	// Run a specific action
	case *model.ActionFlow:
		return runSingleActionFlow(f, event, verbose)
	}

	return nil, nil
}

func RunMultiFlowsRecursive(
	fact *actions.ActionMessage,
	event *ics.VEvent,
	flows model.Flows,
	debugMessages *[]interface{},
	verbose, enableDebugFlow bool,
) error {
	for _, flow := range flows {
		msg, err := RunSingleFlow(event, flow, verbose, enableDebugFlow)
		// oh no, we always exit on errors
		if err != nil {
			return fmt.Errorf("single flow error: %v", err)
		}
		// if msg is null, all good and continue loop
		if msg == nil {
			continue
		}
		switch t := msg.(type) {
		case *ExitFlowsExecutionMessage:
			// exit flow execution loop
			return ErrExited
		case *QueueFlowsExecutionMessage:
			if err = RunMultiFlowsRecursive(fact, event, t.Flows, debugMessages, verbose, enableDebugFlow); err != nil {
				// if a child flow exited (or failed) also exit all parents
				return err
			}
		case *FilterResultExecutionMessage:
			*fact = t.Action
		case *DebugExecutionMessage:
			if enableDebugFlow {
				fmt.Println("[DEBUG]", t.Message)
				*debugMessages = append(*debugMessages, t.Message)
			}
		}
	}
	return nil
}

func (c *ContextFlow) RunMultiFlows(event *ics.VEvent, flows model.Flows) (actions.ActionMessage, error) {
	// filter everything in by default
	var fact actions.ActionMessage = new(actions.FilterInActionMessage)
	err := RunMultiFlowsRecursive(&fact, event, flows, &c.Debugs, c.Verbose, c.EnableDebug)
	return fact, err
}
