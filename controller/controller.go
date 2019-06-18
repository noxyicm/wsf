package controller

import (
	"wsf/config"
	"wsf/controller/dispatcher"
	"wsf/controller/plugin"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/controller/router"
	"wsf/errors"
	"wsf/session"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}

	main Interface
)

// Interface is an interface for controllers
type Interface interface {
	SetRouter(router.Interface) error
	Router() router.Interface
	SetDispatcher(dispatcher.Interface) error
	Dispatcher() dispatcher.Interface
	Dispatch(rqs request.Interface, rsp response.Interface, s session.Interface) error
	Priority() int
	SetThrowExceptions(bool)
	ThrowExceptions() bool
	SetErrorHandling(bool)
	ErrorHandling() bool
	RegisterPlugin(plugin plugin.Interface, priority int) error
	HasPlugin(name string) bool
	GetPlugin(name string) plugin.Interface
}

// Controller base struct
type Controller struct {
	options         *Config
	router          router.Interface
	dispatcher      dispatcher.Interface
	handlers        map[string]func() error
	plugins         *plugin.Broker
	throwExceptions bool
	errorHandling   bool
}

// Init controller resource
func (c *Controller) Init(options *Config) (b bool, err error) {
	c.options = options
	c.dispatcher, err = dispatcher.NewDispatcher(options.Dispatcher.Type, options.Dispatcher)
	if err != nil {
		return false, err
	}

	c.router, err = router.NewRouter(options.Router.Type, options.Router)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Priority returns resource initialization priority
func (c *Controller) Priority() int {
	return c.options.Priority
}

// SetThrowExceptions sets if controller should break on error
func (c *Controller) SetThrowExceptions(v bool) {
	c.throwExceptions = v
}

// ThrowExceptions returns true if controller should break on exception
func (c *Controller) ThrowExceptions() bool {
	return c.throwExceptions
}

// SetErrorHandling sets if a controller should handle errors
func (c *Controller) SetErrorHandling(v bool) {
	c.errorHandling = v
}

// ErrorHandling returns if controller should handle errors
func (c *Controller) ErrorHandling() bool {
	return c.errorHandling
}

// RegisterPlugin registers a plugin to controller
func (c *Controller) RegisterPlugin(p plugin.Interface, priority int) error {
	return c.plugins.Register(p, priority)
}

// HasPlugin returns true if controller has a plugin
func (c *Controller) HasPlugin(name string) bool {
	return c.plugins.Has(name)
}

// GetPlugin returns a plugin by its name
func (c *Controller) GetPlugin(name string) plugin.Interface {
	return c.plugins.Get(name)
}

// NewController creates a new controller specified by type
func NewController(controllerType string, options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[controllerType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized controller type \"%v\"", controllerType)
}

// Register registers a handler for controller creation
func Register(controllerType string, handler func(*Config) (Interface, error)) {
	buildHandlers[controllerType] = handler
}

// SetInstance sets a main controller instance
func SetInstance(ctrl Interface) {
	main = ctrl
}

// Instance returns a main controller instance
func Instance() Interface {
	return main
}

// SetRouter sets controller router
func SetRouter(rtr router.Interface) error {
	return main.SetRouter(rtr)
}

// Router returns controller router
func Router() router.Interface {
	return main.Router()
}

// SetDispatcher sets controller dispatcher
func SetDispatcher(dsp dispatcher.Interface) error {
	return main.SetDispatcher(dsp)
}

// Dispatcher returns controller dispatcher
func Dispatcher() dispatcher.Interface {
	return main.Dispatcher()
}

// SetThrowExceptions sets if controller should break on error
func SetThrowExceptions(b bool) {
	main.SetThrowExceptions(b)
}

// ThrowExceptions returns true if controller should break on exception
func ThrowExceptions() bool {
	return main.ThrowExceptions()
}

// SetErrorHandling sets if a controller should handle errors
func SetErrorHandling(b bool) {
	main.SetErrorHandling(b)
}

// ErrorHandling returns if controller should handle errors
func ErrorHandling() bool {
	return main.ErrorHandling()
}

// RegisterPlugin registers a plugin to controller
func RegisterPlugin(plugin plugin.Interface, priority int) error {
	return main.RegisterPlugin(plugin, priority)
}

// HasPlugin returns true if controller has a plugin
func HasPlugin(name string) bool {
	return main.HasPlugin(name)
}

// GetPlugin returns a plugin by its name
func GetPlugin(name string) plugin.Interface {
	return main.GetPlugin(name)
}
