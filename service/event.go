package service

// Event is a service specific event
type Event interface {
	Error() error
	Message() string
}

// Debug is a simple string event
type Debug struct {
	msg string
}

// Error returns an event error
func (e *Debug) Error() error {
	return nil
}

// Message returns an event message
func (e *Debug) Message() string {
	return e.msg
}

// Info is a simple string event
type Info struct {
	msg string
}

// Error returns an event error
func (e *Info) Error() error {
	return nil
}

// Message returns an event message
func (e *Info) Message() string {
	return e.msg
}

// Error is a simple string event
type Error struct {
	err error
}

// Error returns an event error
func (e *Error) Error() error {
	return e.err
}

// Message returns an event message
func (e *Error) Message() string {
	return e.err.Error()
}

// DebugEvent creates a Debug event
func DebugEvent(msg string) Event {
	return &Debug{msg: msg}
}

// InfoEvent creates an Info event
func InfoEvent(msg string) Event {
	return &Info{msg: msg}
}

// ErrorEvent creates an Error event
func ErrorEvent(err error) Event {
	return &Error{err: err}
}
