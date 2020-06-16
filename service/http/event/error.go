package event

import (
	"net/http"
	"time"
)

// Error represents singular http error event
type Error struct {
	Request *http.Request
	err     error
	start   time.Time
	elapsed time.Duration
}

// Error service.Event interface implementation
func (e *Error) Error() error {
	return e.err
}

// Message service.Event interface implementation
func (e *Error) Message() string {
	return e.err.Error()
}

// Elapsed returns duration of the invocation
func (e *Error) Elapsed() time.Duration {
	return e.elapsed
}

// NewError creates new error event
func NewError(r *http.Request, err error, start time.Time) *Error {
	return &Error{Request: r, err: err, start: start, elapsed: time.Since(start)}
}
