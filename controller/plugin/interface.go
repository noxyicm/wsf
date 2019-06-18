package plugin

import (
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/session"
)

// Interface represents controller plugin
type Interface interface {
	Name() string
	RouteStartup(rqs request.Interface, rsp response.Interface, s session.Interface) (bool, error)
	RouteShutdown(rqs request.Interface, rsp response.Interface, s session.Interface) (bool, error)
	DispatchLoopStartup(rqs request.Interface, rsp response.Interface, s session.Interface) (bool, error)
	PreDispatch(rqs request.Interface, rsp response.Interface, s session.Interface) (bool, error)
	PostDispatch(rqs request.Interface, rsp response.Interface, s session.Interface) (bool, error)
	DispatchLoopShutdown(rqs request.Interface, rsp response.Interface, s session.Interface) (bool, error)
}

// ControllerWithExceptionInterface is a wrapper for main controller
type ControllerWithExceptionInterface interface {
	SetThrowExceptions(bool)
	ThrowExceptions() bool
	ErrorHandling()
}
