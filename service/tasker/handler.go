package tasker

import (
	"wsf/context"
	"wsf/errors"
)

// Handler represents a worker handler interface
type Handler interface {
	Waiter

	New(*Task) (Handler, error)
	StartRoutine(ctx context.Context, task *Task) error
	Start(ctx context.Context)
	Stop()
	Name() string
	TaskID() int64
	InChannel() (chan<- *Message, error)
}

// HandleTask creates a new instance of task specific handler and runs a routine for it
func HandleTask(ctx context.Context, task *Task) (Handler, error) {
	if !HasHandler(task.Handler) {
		return nil, errors.Errorf("Handler by name '%s' is not registered", task.Handler)
	}

	hndlr, err := NewHandler(task.Handler, task)
	if err != nil {
		return nil, err
	}

	go hndlr.Start(ctx)
	return hndlr, nil
}
