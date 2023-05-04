package engine

import (
	"github.com/ralf-life/engine/actions"
	"github.com/ralf-life/engine/model"
)

// ExecutionMessage specifies what should happen next
type ExecutionMessage interface {
}

// ExitFlowsExecutionMessage exists all flows
type ExitFlowsExecutionMessage struct {
}

// QueueFlowsExecutionMessage adds more flows to execute
type QueueFlowsExecutionMessage struct {
	Flows model.Flows
}

type FilterResultExecutionMessage struct {
	Action actions.ActionMessage
}

type DebugExecutionMessage struct {
	Message interface{}
}
