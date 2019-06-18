package errors

import (
	"fmt"
	"io"
)

// HTTPError is a http error
type HTTPError struct {
	msg   string
	code  int
	cause error
	*stack
}

// Error returns error message
func (e *HTTPError) Error() string {
	return e.msg
}

// Code returns error code
func (e *HTTPError) Code() int {
	return e.code
}

// Format formats error
func (e *HTTPError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, e.msg)
			if e.cause != nil {
				io.WriteString(s, ": "+e.cause.Error())
			}

			e.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		if s.Flag('+') {
			io.WriteString(s, e.msg)
			if e.cause != nil {
				io.WriteString(s, ": "+e.cause.Error())
			}

			e.stack.Format(s, verb)
			return
		}

		io.WriteString(s, e.msg)
		if e.cause != nil {
			io.WriteString(s, ": "+e.cause.Error())
		}
	case 'q':
		if e.cause != nil {
			fmt.Fprintf(s, "%q", e.msg+": "+e.cause.Error())
			return
		}

		fmt.Fprintf(s, "%q", e.msg)
	}
}

// NewHTTP returns an error with the supplied message
func NewHTTP(message string, code int) error {
	return &HTTPError{
		msg:   message,
		code:  code,
		stack: callers(),
	}
}

// ErrorHTTPf returns an error annotating err with a stack trace
// at the point ErrorHTTPf is called, and the format specifier
// If err is nil, ErrorHTTPf returns nil
func ErrorHTTPf(format string, code int, args ...interface{}) error {
	return &HTTPError{
		msg:   fmt.Sprintf(format, args...),
		code:  code,
		stack: callers(),
	}
}

// WrapHTTP returns an error annotating err with a stack trace
// at the point WrapHTTP is called, and the supplied message
// If err is nil, Wrap returns nil
func WrapHTTP(err error, message string, code int) error {
	if err == nil {
		return nil
	}

	return &HTTPError{
		msg:   message,
		code:  code,
		cause: err,
		stack: callers(),
	}
}

// WrapHTTPf returns an error annotating err with a stack trace
// at the point WrapHTTPf is called, and the format specifier
// If err is nil, Wrapf returns nil
func WrapHTTPf(err error, format string, code int, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &HTTPError{
		msg:   fmt.Sprintf(format, args...),
		code:  code,
		cause: err,
		stack: callers(),
	}
}
