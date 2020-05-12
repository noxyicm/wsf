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

	// MessageReport indicates that task has something to say
	MessageReport = 0

	// MessageAddTask indicates that task must be added to queue
	MessageAddTask = 1

	// MessageTaskAdded indicates that task has been added to queue
	MessageTaskAdded = 2

	// MessageStartTask indicates that task must be started
	MessageStartTask = 3

	// MessageTaskStarted indicates that task has been started
	MessageTaskStarted = 4

	// MessageModifyTask indicates that task must be modified
	MessageModifyTask = 5

	// MessageTaskModified indicates that task has been modified
	MessageTaskModified = 6

	// MessageStopTask indicates that task must be stoped
	MessageStopTask = 7

	// MessageTaskStoped indicates that task has been stoped
	MessageTaskStoped = 8

	// MessageRemoveTask indicates that task must be removed from queue
	MessageRemoveTask = 9

	// MessageTaskRemoved indicates that task has been removed from queue
	MessageTaskRemoved = 10

	// MessageTaskDone indicates that task has been done
	MessageTaskDone = 11

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
