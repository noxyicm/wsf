package controller

import (
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
)

const (
	// TYPEDfault is a type of controller
	TYPEDfault = "default"
)

func init() {
	Register(TYPEDfault, NewDefaultController)
}

// Default is a default controller for dc
type Default struct {
	Controller
}

// AddActionController adds an action controller constructor
func (c *Default) AddActionController(moduleName string, controllerName string, cnstr func() (ActionControllerInterface, error)) error {
	return c.dispatcher.AddActionController(moduleName, controllerName, cnstr)
}

// SetRouter sets controller router
func (c *Default) SetRouter(r RouterInterface) error {
	c.router = r
	return nil
}

// Router returns controller router
func (c *Default) Router() RouterInterface {
	return c.router
}

// SetDispatcher sets controller dispatcher
func (c *Default) SetDispatcher(d DispatcherInterface) error {
	c.dispatcher = d
	return nil
}

// Dispatcher returns controller dispatcher
func (c *Default) Dispatcher() DispatcherInterface {
	return c.dispatcher
}

// Dispatch dispatches the reauest into the dispatcher loop
func (c *Default) Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) error {
	if c.ErrorHandling() && !c.plugins.Has("ErrorHandler") {
		p, err := NewErrorHandlerPlugin()
		if err != nil {
			return err
		}

		c.plugins.Register(p, 100)
	}

	ok, err := c.plugins.RouteStartup(ctx, rqs, rsp)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	ok, err = c.router.Match(ctx, rqs)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	ok, err = c.plugins.RouteShutdown(ctx, rqs, rsp)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	ok, err = c.plugins.DispatchLoopStartup(ctx, rqs, rsp)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	for {
		if rqs.IsDispatched() {
			goto done
		}

		rqs.SetDispatched(true)

		// Notify plugins of dispatch startup
		ok, err = c.plugins.PreDispatch(ctx, rqs, rsp)
		if !ok && c.ThrowExceptions() {
			return err
		} else if err != nil {
			rsp.SetException(err)
		}

		// Skip requested action if PreDispatch() has reset it
		if !rqs.IsDispatched() {
			continue
		}

		// Dispatch request
		ok, err = c.dispatcher.Dispatch(ctx, rqs, rsp)
		if !ok && c.ThrowExceptions() {
			return err
		} else if err != nil {
			rsp.SetException(err)
		}

		// Notify plugins of dispatch completion
		ok, err = c.plugins.PostDispatch(ctx, rqs, rsp)
		if !ok && c.ThrowExceptions() {
			return err
		} else if err != nil {
			rsp.SetException(err)
		}
	}

done:
	// Notify plugins of dispatch loop completion
	ok, err = c.plugins.DispatchLoopShutdown(ctx, rqs, rsp)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	return nil
}

// NewDefaultController creates new default controller
func NewDefaultController(cfg *Config) (ci Interface, err error) {
	c := &Default{}
	c.Options = cfg
	c.handlers = make(map[string]func() error)
	c.plugins, err = NewPluginBroker()
	if err != nil {
		return nil, err
	}

	c.helpers, err = NewHelperBroker()
	if err != nil {
		return nil, err
	}

	c.errorHandling = cfg.ErrorHandling
	c.throwExceptions = cfg.ThrowExceptions

	return c, nil
}
