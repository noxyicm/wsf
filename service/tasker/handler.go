package tasker

import "context"

// Handler represents a worker handler interface
type Handler interface {
	StartRoutine(ctx context.Context, task *Task, outChan chan TaskState)
}
