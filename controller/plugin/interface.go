package plugin

import (
	"wsf/controller/context"
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
