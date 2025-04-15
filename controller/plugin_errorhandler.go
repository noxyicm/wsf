package controller

import (
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
)

const (
	// TYPEControllerPluginTypeErrorHandler name of the type of the plugin
	TYPEControllerPluginTypeErrorHandler = "ErrorHandler"
)

func init() {
	RegisterPluginType(TYPEControllerPluginTypeErrorHandler, NewErrorHandlerPlugin)
}

// ErrorHandler is a plugin for handling errors
type ErrorHandler struct {
	name                           string
	module                         string
	controller                     string
	action                         string
	handleErrors                   bool
	isInsideErrorHandlerLoop       bool
	exceptionCountAtFirstEncounter int
}

// Name returns plugin name
func (p *ErrorHandler) Name() string {
	return p.name
}

// SetModule sets plugin module
func (p *ErrorHandler) SetModule(md string) {
	p.module = md
}

// Module returns plugin module
func (p *ErrorHandler) Module() string {
	return p.module
}

// SetController sets plugin controller
func (p *ErrorHandler) SetController(ctrl string) {
	p.controller = ctrl
}

// Controller returns plugin controller
func (p *ErrorHandler) Controller() string {
	return p.controller
}

// SetAction sets plugin action
func (p *ErrorHandler) SetAction(action string) {
	p.action = action
}

// Action returns plugin action
func (p *ErrorHandler) Action() string {
	return p.action
}

// SetErrorHandlerModule sets routing module for request
func (p *ErrorHandler) SetErrorHandlerModule(value string) error {
	p.module = value
	return nil
}

// ErrorHandlerModule returns routing module for request
func (p *ErrorHandler) ErrorHandlerModule() string {
	return p.module
}

// SetErrorHandlerController sets routing controller for request
func (p *ErrorHandler) SetErrorHandlerController(value string) error {
	p.controller = value
	return nil
}

// ErrorHandlerController returns routing controller for request
func (p *ErrorHandler) ErrorHandlerController() string {
	return p.controller
}

// SetErrorHandlerAction sets routing action for request
func (p *ErrorHandler) SetErrorHandlerAction(value string) error {
	p.action = value
	return nil
}

// ErrorHandlerAction returns routing action for request
func (p *ErrorHandler) ErrorHandlerAction() string {
	return p.action
}

// RouteStartup routine
func (p *ErrorHandler) RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return p.handleError(ctx, rqs, rsp)
}

// RouteShutdown routine
func (p *ErrorHandler) RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return p.handleError(ctx, rqs, rsp)
}

// DispatchLoopStartup routine
func (p *ErrorHandler) DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return p.handleError(ctx, rqs, rsp)
}

// PreDispatch routine
func (p *ErrorHandler) PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return p.handleError(ctx, rqs, rsp)
}

// PostDispatch routine
func (p *ErrorHandler) PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return p.handleError(ctx, rqs, rsp)
}

// DispatchLoopShutdown routine
func (p *ErrorHandler) DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return p.handleError(ctx, rqs, rsp)
}

// SetHandleErrors sets if errors should be handled
func (p *ErrorHandler) SetHandleErrors(v bool) {
	p.handleErrors = v
}

func (p *ErrorHandler) handleError(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	if !p.handleErrors {
		return true, nil
	}

	encounteredError := rqs.Param(p.name)
	if encounteredError != nil {
		if len(rsp.Exceptions()) > encounteredError.(*errors.Exception).Encountered {
			return false, rsp.Exceptions()[len(rsp.Exceptions())-1]
		}
	}

	// check for an exception AND allow the error handler controller the option to forward
	if rsp.IsException() && encounteredError == nil {
		// Get exception information
		exceptions := rsp.Exceptions()
		exception := exceptions[0]
		err := errors.NewException(exception)

		// Keep a copy of the original request
		orqs := rqs
		err.Request = orqs
		err.Encountered = len(exceptions)

		// Forward to the error handler
		rqs.SetParam(p.name, err)
		rqs.SetModuleName(p.ErrorHandlerModule())
		rqs.SetControllerName(p.ErrorHandlerController())
		rqs.SetActionName(p.ErrorHandlerAction())
		rqs.SetDispatched(false)
	}

	return true, nil
}

// NewErrorHandlerPlugin creates a new error handling plugin
func NewErrorHandlerPlugin(name string) (PluginInterface, error) {
	return &ErrorHandler{
		name:         name,
		module:       "index",
		controller:   "error",
		action:       "error",
		handleErrors: true,
	}, nil
}
