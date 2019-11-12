package errors

import (
	"fmt"
	"wsf/controller/request"
)

// Exception constants
const (
	ExceptionOther = iota
	ExceptionNoRoute
	ExceptionNoController
	ExceptionNoAction
)

// ControllerRouterException is a router exception
type ControllerRouterException struct {
	msg  string
	code int
}

// Error returns exception message
func (e *ControllerRouterException) Error() string {
	return e.msg
}

// Code returns exception code
func (e *ControllerRouterException) Code() int {
	return e.code
}

// ControllerDispatcherException is a dispatcher exception
type ControllerDispatcherException struct {
	msg string
}

// Error returns exception message
func (e *ControllerDispatcherException) Error() string {
	return e.msg
}

// ControllerActionException is an action exception
type ControllerActionException struct {
	msg  string
	code int
}

// Error returns exception message
func (e *ControllerActionException) Error() string {
	return e.msg
}

// Code returns exception code
func (e *ControllerActionException) Code() int {
	return e.code
}

// Exception represents an exception
type Exception struct {
	Typ         int
	Request     request.Interface
	Encountered int
	Original    error
}

// Error returns exception message
func (e *Exception) Error() string {
	return e.Original.Error()
}

// WithTrace returns exception message with stack trace
func (e *Exception) WithTrace() string {
	return fmt.Sprintf("%+v", e.Original)
}

// Code returns exception code
func (e *Exception) Code() int {
	switch e.Original.(type) {
	case *HTTPError:
		return e.Original.(*HTTPError).Code()
	}

	return 500
}

// NewException creates a new exception
func NewException(err error) *Exception {
	return &Exception{
		Original: err,
	}
}
