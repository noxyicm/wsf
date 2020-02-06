package tasker

import (
	"wsf/context"
	"wsf/log"
)

const (
	// TYPEDefaultWorker is a type id of default worker
	TYPEDefaultWorker = "default"

	// TaskStatusReady represents task that is ready do work
	TaskStatusReady = 1
	// TaskStatusInProgress represents task that is in work
	TaskStatusInProgress = 100
	// TaskStatusOver represents task that is over
	TaskStatusOver = 200
	// TaskStatusFail represents task that is failed
	TaskStatusFail = 500
)

// Worker is a worker
type Worker interface {
	Waiter

	New() (Worker, error)
	Start(ctx context.Context) error
	Stop()
	SetLogger(*log.Log) error
}

// Waiter interface
type Waiter interface {
	ID() int64
	Wait() <-chan Message
}

// Message is simple message struct for reporting throught Waiter channels
type Message struct {
	Error    error
	Text     string
	Priority int
}
