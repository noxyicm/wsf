package controller

import (
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
)

var (
	buildPluginHandlers = map[string]func(string) (PluginInterface, error){}
)

// PluginInterface represents controller plugin
type PluginInterface interface {
	Name() string
	RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
}

// WithExceptionInterface is a wrapper for main controller
type WithExceptionInterface interface {
	SetThrowExceptions(bool)
	ThrowExceptions() bool
	ErrorHandling()
}

// NewPluginType creates a new controller plugin from given type and options
func NewPlugin(pluginType string, pluginName string) (pi PluginInterface, err error) {
	if f, ok := buildPluginHandlers[pluginType]; ok {
		return f(pluginName)
	}

	return nil, errors.Errorf("Unrecognized controller plugin type \"%v\"", pluginType)
}

// RegisterPluginType registers a handler for controller plugin creation
func RegisterPluginType(pluginType string, handler func(string) (PluginInterface, error)) {
	buildPluginHandlers[pluginType] = handler
}

// PluginAbstract is a extendable plugin base
type PluginAbstract struct {
	name string
}

// Name returns plugin name
func (p *PluginAbstract) Name() string {
	return p.name
}

// RouteStartup routine
func (p *PluginAbstract) RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// RouteShutdown routine
func (p *PluginAbstract) RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// DispatchLoopStartup routine
func (p *PluginAbstract) DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// PreDispatch routine
func (p *PluginAbstract) PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// PostDispatch routine
func (p *PluginAbstract) PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// DispatchLoopShutdown routine
func (p *PluginAbstract) DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// NewPluginAbstract creates a plugin abstract instance
func NewPluginAbstract(name string) (*PluginAbstract, error) {
	return &PluginAbstract{
		name: name,
	}, nil
}
