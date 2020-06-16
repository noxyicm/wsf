package plugin

import (
	"wsf/context"
	"wsf/controller/request"
	"wsf/controller/response"
)

// Interface represents controller plugin
type Interface interface {
	Name() string
	RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
}

// ControllerWithExceptionInterface is a wrapper for main controller
type ControllerWithExceptionInterface interface {
	SetThrowExceptions(bool)
	ThrowExceptions() bool
	ErrorHandling()
}

// Abstract is a extendable plugin base
type Abstract struct {
	name string
}

// Name returns plugin name
func (p *Abstract) Name() string {
	return p.name
}

// RouteStartup routine
func (p *Abstract) RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// RouteShutdown routine
func (p *Abstract) RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// DispatchLoopStartup routine
func (p *Abstract) DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// PreDispatch routine
func (p *Abstract) PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// PostDispatch routine
func (p *Abstract) PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// DispatchLoopShutdown routine
func (p *Abstract) DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// NewAbstract creates a new abstract plugin
func NewAbstract(name string) *Abstract {
	return &Abstract{
		name: name,
	}
}
