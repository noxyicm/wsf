package controller

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/utils"
)

var (
	buildRouterHandlers = map[string]func(config.Config) (RouterInterface, error){}
)

// RouterInterface is an interface for controllers
type RouterInterface interface {
	AddRoute(routeType string, route string, defaults map[string]string, reqs map[string]string, name string) error
	RouteByName(name string) (RouteInterface, bool)
	Match(context.Context, request.Interface) (bool, error)
	Assemble(ctx context.Context, params map[string]interface{}, name string, reset bool, encode bool) (string, error)
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
func NewRouter(routerType string, options config.Config) (RouterInterface, error) {
	if f, ok := buildRouterHandlers[routerType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized router type \"%v\"", routerType)
}

// RegisterRouter registers a handler for router creation
func RegisterRouter(routerType string, handler func(config.Config) (RouterInterface, error)) {
	buildRouterHandlers[routerType] = handler
}
