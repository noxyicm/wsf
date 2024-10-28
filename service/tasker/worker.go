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

	// MessageTaskNotStarted indicates that task has not been started
	MessageTaskNotStarted = 13

	// MessageModifyTask indicates that task must be modified
	MessageModifyTask = 5

	// MessageTaskModified indicates that task has been modified
	MessageTaskModified = 6

	// MessageTaskNotModified indicates that task has not been modified
	MessageTaskNotModified = 12

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

	// MessageWorkerStart indicates that worker must be started
	MessageWorkerStart = 100

	// MessageWorkerStarted indicates that worker started
	MessageWorkerStarted = 101

	// MessageWorkerStop indicates that worker must be stoped
	MessageWorkerStop = 190

	// MessageWorkerStoped indicates that worker stoped
	MessageWorkerStoped = 191

	// ScopeGlobal indicates that message is visible globaly
	ScopeGlobal = 1
	// ScopeTasker indicates that message is visible only for tasker and lower
	ScopeTasker = 2
	// ScopeWorker indicates that message is visible only for worker and lower
	ScopeWorker = 3
	// ScopeHandler indicates that message is visible only for handler
	ScopeHandler = 4
)

// Worker is a worker
type Worker interface {
	Waiter

	New() (Worker, error)
	Start(ctx context.Context) error
	StartTask(ctx context.Context, tsk *Task) error
	StartHandler(tsk *Task) error
	Stop()
	SetLogger(*log.Log) error
	IsPersistent() bool
	IsWorking() bool
	IsAutoStart() bool
	CanHandleMore() bool
	CanReceiveTasks() bool
	InChannel() (chan<- *Message, error)
	Handler(name string, indx int) (chan<- *Message, error)
	HandlerWithTask(name string, taskID int64) (chan<- *Message, error)
	HandlerInstanceWithTask(name string, taskID int64) (Handler, error)
	HandlerInstances(name string) ([]Handler, error)
}

// Waiter interface
type Waiter interface {
	ID() int64
	Wait() <-chan *Message
}

// Message is simple message struct for comunicationg through channels
type Message struct {
	ID       int64
	Type     int
	Error    error
	Text     string
	Priority int
	Scope    int
	Task     Task
	Previous *Message
}

// IsReplyTo returns true if this message is a reply to message with provided ID
func (m *Message) IsReplyTo(id int64) bool {
	if m.Previous != nil && m.Previous.ID == id {
		return true
	}

	return false
}
