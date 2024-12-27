package http

import (
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
)

var (
	buildMiddlewareHandlers = map[string]func(*MiddlewareConfig) (Middleware, error){}
)

// Middleware for reauest manipulation
// type Middleware func(f http.HandlerFunc) http.HandlerFunc
type Middleware interface {
	Init(options *MiddlewareConfig) (bool, error)
	Handle(s *Service, r request.Interface, w response.Interface) bool
}

// NewMiddleware creates a new middleware specified by type
func NewMiddleware(middlewareType string, options *MiddlewareConfig) (Middleware, error) {
	if f, ok := buildMiddlewareHandlers[middlewareType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized middleware type \"%v\"", middlewareType)
}

// RegisterMiddleware registers a handler for middleware creation
func RegisterMiddleware(middlewareType string, handler func(*MiddlewareConfig) (Middleware, error)) {
	buildMiddlewareHandlers[middlewareType] = handler
}
