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

	// MessageAddTask type 1
	MessageAddTask = 1

	// MessageModifyTask type 2
	MessageModifyTask = 2

	// MessageStopTask type 3
	MessageStopTask = 3

	// MessageReport type 11
	MessageReport = 11

	// MessageWorkerStart asd
	MessageWorkerStart = 100

	// MessageWorkerStop asd
	MessageWorkerStop = 190
)

// Worker is a worker
type Worker interface {
	Waiter

	New() (Worker, error)
	Start(ctx context.Context) error
	Stop()
	SetLogger(*log.Log) error
	IsWorking() bool
	IsAutoStart() bool
	CanHandleMore() bool
	CanReceiveTasks() bool
	InChannel() (chan<- *Message, error)
	Handler(name string, indx int) (chan<- *Message, error)
}

// Waiter interface
type Waiter interface {
	ID() int64
	Wait() <-chan *Message
}

// Message is simple message struct for comunicationg through channels
type Message struct {
	Type     int
	Error    error
	Text     string
	Priority int
	Task     Task
}
