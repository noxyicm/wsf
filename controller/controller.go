package controller

import (
	"wsf/config"
	"wsf/context"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}

	main Interface
)

// Interface is an interface for controllers
type Interface interface {
	SetRouter(RouterInterface) error
	Router() RouterInterface
	SetDispatcher(DispatcherInterface) error
	Dispatcher() DispatcherInterface
	Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) error
	AddActionController(moduleName string, controllerName string, cnstr func() (ActionControllerInterface, error)) error
	Priority() int
	SetThrowExceptions(bool)
	ThrowExceptions() bool
	SetErrorHandling(bool)
	ErrorHandling() bool
	RegisterPlugin(plugin PluginInterface, priority int) error
	HasPlugin(name string) bool
	Plugin(name string) PluginInterface
	SetHelperBroker(broker *HelperBroker) error
	HelperBroker() *HelperBroker
	HasHelper(name string) bool
	Helper(name string) HelperInterface
}

// Controller base struct
type Controller struct {
	Options         *Config
	router          RouterInterface
	dispatcher      DispatcherInterface
	handlers        map[string]func() error
	plugins         *PluginBroker
	helpers         *HelperBroker
	throwExceptions bool
	errorHandling   bool
}

// Init controller resource
func (c *Controller) Init(options *Config) (b bool, err error) {
	c.Options = options
	c.dispatcher, err = NewDispatcher(options.Dispatcher.Type, options.Dispatcher)
	if err != nil {
		return false, err
	}

	c.router, err = NewRouter(options.Router.GetString("type"), options.Router)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Priority returns resource initialization priority
func (c *Controller) Priority() int {
	return c.Options.Priority
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
func (c *Controller) RegisterPlugin(p PluginInterface, priority int) error {
	return c.plugins.Register(p, priority)
}

// HasPlugin returns true if controller has a plugin
func (c *Controller) HasPlugin(name string) bool {
	return c.plugins.Has(name)
}

// Plugin returns a plugin by its name
func (c *Controller) Plugin(name string) PluginInterface {
	return c.plugins.Get(name)
}

// SetHelperBroker sets helper broker
func (c *Controller) SetHelperBroker(broker *HelperBroker) (err error) {
	if broker != nil {
		c.helpers = broker
	} else {
		c.helpers, err = NewHelperBroker()
		if err != nil {
			return err
		}
	}

	return nil
}

// HelperBroker returns action controller helper broker
func (c *Controller) HelperBroker() *HelperBroker {
	return c.helpers
}

// HasHelper returns true if action Halper is registered
func (c *Controller) HasHelper(name string) bool {
	return c.helpers.HasHelper(name)
}

// Helper returns action Halper
func (c *Controller) Helper(name string) HelperInterface {
	h, _ := c.helpers.GetHelper(name)
	return h
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
func SetRouter(rtr RouterInterface) error {
	return main.SetRouter(rtr)
}

// Router returns controller router
func Router() RouterInterface {
	return main.Router()
}

// SetDispatcher sets controller dispatcher
func SetDispatcher(dsp DispatcherInterface) error {
	return main.SetDispatcher(dsp)
}

// Dispatcher returns controller dispatcher
func Dispatcher() DispatcherInterface {
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
func RegisterPlugin(plugin PluginInterface, priority int) error {
	return main.RegisterPlugin(plugin, priority)
}

// HasPlugin returns true if controller has a plugin
func HasPlugin(name string) bool {
	return main.HasPlugin(name)
}

// Plugin returns a plugin by its name
func Plugin(name string) PluginInterface {
	return main.Plugin(name)
}

// GetHelperBroker returns controller helper broker
func GetHelperBroker() *HelperBroker {
	return main.HelperBroker()
}

// HasHelper returns true if controller has a helper
func HasHelper(name string) bool {
	return main.HasHelper(name)
}

// Helper returns a helper by its name
func Helper(name string) HelperInterface {
	return main.Helper(name)
}
