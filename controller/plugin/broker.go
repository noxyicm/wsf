package plugin

import (
	"wsf/context"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/utils/stack"
)

// Broker handles dispatching of events to plugins
type Broker struct {
	plugins *stack.Prioritised
}

// Register a plugin
func (b *Broker) Register(plugin Interface, priority int) error {
	if b.plugins.Contains(plugin) {
		return errors.Errorf("Plugin already registered")
	}

	b.plugins.Push(priority, plugin)
	return nil
}

// Has returns true if plugin is already registered
func (b *Broker) Has(pluginName string) bool {
	for _, p := range b.Plugins() {
		if p.Name() == pluginName {
			return true
		}
	}

	return false
}

// Get returns plugin by its name if plugin is already registered
func (b *Broker) Get(pluginName string) Interface {
	for _, p := range b.Plugins() {
		if p.Name() == pluginName {
			return p
		}
	}

	return nil
}

// Plugins returns plugins stack
func (b *Broker) Plugins() []Interface {
	stk := make([]Interface, b.plugins.Len())
	for k, p := range b.plugins.Stack() {
		stk[k] = p.(Interface)
	}

	return stk
}

// RouteStartup notifyes plugins of RouteStartup routine
func (b *Broker) RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	for _, plugin := range b.Plugins() {
		if ok, err = plugin.RouteStartup(ctx, rqs, rsp); !ok {
			return false, err
		}
	}

	return true, err
}

// RouteShutdown notifyes plugins of RouteShutdown routine
func (b *Broker) RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	for _, plugin := range b.Plugins() {
		if ok, err = plugin.RouteShutdown(ctx, rqs, rsp); !ok {
			return false, err
		}
	}

	return true, err
}

// DispatchLoopStartup notifyes plugins of DispatchLoopStartup routine
func (b *Broker) DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	for _, plugin := range b.Plugins() {
		if ok, err = plugin.DispatchLoopStartup(ctx, rqs, rsp); !ok {
			return false, err
		}
	}

	return true, err
}

// PreDispatch notifyes plugins of PreDispatch routine
func (b *Broker) PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	for _, plugin := range b.Plugins() {
		if ok, err = plugin.PreDispatch(ctx, rqs, rsp); !ok {
			return false, err
		}
	}

	return true, err
}

// PostDispatch notifyes plugins of PostDispatch routine
func (b *Broker) PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	for _, plugin := range b.Plugins() {
		if ok, err = plugin.PostDispatch(ctx, rqs, rsp); !ok {
			return false, err
		}
	}

	return true, err
}

// DispatchLoopShutdown notifyes plugins of DispatchLoopShutdown routine
func (b *Broker) DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	for _, plugin := range b.Plugins() {
		if ok, err = plugin.DispatchLoopShutdown(ctx, rqs, rsp); !ok {
			return false, err
		}
	}

	return true, err
}

// NewBroker creates a new plugins broker
func NewBroker() (*Broker, error) {
	return &Broker{
		plugins: stack.NewPrioritised(),
	}, nil
}
