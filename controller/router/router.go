package router

import (
	"wsf/config"
	"wsf/controller/request"
	"wsf/errors"
	"wsf/utils"
)

var (
	buildHandlers = map[string]func(config.Config) (Interface, error){}
)

// Interface is an interface for controllers
type Interface interface {
	AddRoute(routeType string, route string, defaults map[string]string, reqs map[string]string, name string) error
	Route(request.Interface) (bool, error)
}

// ReverseRoutes reorders routes map
func ReverseRoutes(m map[string]RouteInterface) map[string]RouteInterface {
	s := make(map[string]RouteInterface)
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	keys = utils.ReverseSliceS(keys)
	for _, k := range keys {
		s[k] = m[k]
	}

	return s
}

// NewRouter creates a new router specified by type
func NewRouter(routerType string, options config.Config) (Interface, error) {
	if f, ok := buildHandlers[routerType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized router type \"%v\"", routerType)
}

// Register registers a handler for router creation
func Register(routerType string, handler func(config.Config) (Interface, error)) {
	buildHandlers[routerType] = handler
}
