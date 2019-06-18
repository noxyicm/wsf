package tasker

import "time"

// Task represents task
type Task struct {
	ID         int
	State      int
	Name       string
	Date       time.Time
	Year       int
	Month      int
	Day        int
	Hour       int
	Minute     int
	Second     int
	Interval   bool
	Handler    string
	Dataset    string
	LastID     int
	LastUpdate time.Time
	Datahash   string
}

// TaskState represents item state
type TaskState struct {
	ID    int
	State bool
	Error error
}
