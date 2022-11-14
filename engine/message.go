package engine

import "github.com/ralf-life/engine/model"

// ExecutionMessage specifies what should happen next
type ExecutionMessage interface {
}

// ExitMessage exists all flows
type ExitMessage struct {
}

// QueueMessage adds more flows to execute
type QueueMessage struct {
	Flows model.Flows
}
