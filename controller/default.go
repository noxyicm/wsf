package controller

import (
	"wsf/controller/dispatcher"
	"wsf/controller/plugin"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/controller/router"
	"wsf/session"
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

// SetRouter sets controller router
func (c *Default) SetRouter(r router.Interface) error {
	c.router = r
	return nil
}

// Router returns controller router
func (c *Default) Router() router.Interface {
	return c.router
}

// SetDispatcher sets controller dispatcher
func (c *Default) SetDispatcher(d dispatcher.Interface) error {
	c.dispatcher = d
	return nil
}

// Dispatcher returns controller dispatcher
func (c *Default) Dispatcher() dispatcher.Interface {
	return c.dispatcher
}

// Dispatch dispatches the reauest into the dispatcher loop
func (c *Default) Dispatch(rqs request.Interface, rsp response.Interface, s session.Interface) error {
	if c.ErrorHandling() && !c.plugins.Has("ErrorHandler") {
		p, err := plugin.NewErrorHandler()
		if err != nil {
			return err
		}

		c.plugins.Register(p, 100)
	}

	ok, err := c.plugins.RouteStartup(rqs, rsp, s)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	ok, err = c.router.Route(rqs)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	ok, err = c.plugins.RouteShutdown(rqs, rsp, s)
	if !ok && c.ThrowExceptions() {
		return err
	} else if err != nil {
		rsp.SetException(err)
	}

	ok, err = c.plugins.DispatchLoopStartup(rqs, rsp, s)
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
		ok, err = c.plugins.PreDispatch(rqs, rsp, s)
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
		ok, err = c.dispatcher.Dispatch(rqs, rsp, s)
		if !ok && c.ThrowExceptions() {
			return err
		} else if err != nil {
			rsp.SetException(err)
		}

		// Notify plugins of dispatch completion
		ok, err = c.plugins.PostDispatch(rqs, rsp, s)
		if !ok && c.ThrowExceptions() {
			return err
		} else if err != nil {
			rsp.SetException(err)
		}
	}

done:
	// Notify plugins of dispatch loop completion
	ok, err = c.plugins.DispatchLoopShutdown(rqs, rsp, s)
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
	c.options = cfg
	c.handlers = make(map[string]func() error)
	c.plugins, err = plugin.NewBroker()
	c.errorHandling = cfg.ErrorHandling
	c.throwExceptions = cfg.ThrowExceptions
	if err != nil {
		return nil, err
	}

	return c, nil
}
