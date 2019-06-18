package event

import (
	"net/http"
	"time"
)

// Error represents singular http error event
type Error struct {
	Request *http.Request
	Error   error
	start   time.Time
	elapsed time.Duration
}

// Elapsed returns duration of the invocation
func (e *Error) Elapsed() time.Duration {
	return e.elapsed
}

// NewError creates new error event
func NewError(r *http.Request, err error, start time.Time) *Error {
	return &Error{Request: r, Error: err, start: start, elapsed: time.Since(start)}
}
