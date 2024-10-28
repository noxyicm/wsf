package controller

import (
	"wsf/context"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/utils/stack"
)

// PluginBroker handles dispatching of events to plugins
type PluginBroker struct {
	plugins *stack.Prioritised
}

// Register a plugin
func (b *PluginBroker) Register(plugin PluginInterface, priority int) error {
	if b.plugins.Contains(plugin) {
		return errors.Errorf("Plugin already registered")
	}

	b.plugins.Push(priority, plugin)
	return nil
}

// Has returns true if plugin is already registered
func (b *PluginBroker) Has(pluginName string) bool {
	for _, p := range b.Plugins() {
		if p.Name() == pluginName {
			return true
		}
	}

	return false
}

// Get returns plugin by its name if plugin is already registered
func (b *PluginBroker) Get(pluginName string) PluginInterface {
	for _, p := range b.Plugins() {
		if p.Name() == pluginName {
			return p
		}
	}

	return nil
}

// Plugins returns plugins stack
func (b *PluginBroker) Plugins() []PluginInterface {
	stk := make([]PluginInterface, b.plugins.Len())
	for k, p := range b.plugins.Stack() {
		stk[k] = p.(PluginInterface)
	}

	return stk
}

// RouteStartup notifyes plugins of RouteStartup routine
func (b *PluginBroker) RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	defer b.recover(ctx, rqs, rsp)

	for _, plugin := range b.Plugins() {
		if ok, err = plugin.RouteStartup(ctx, rqs, rsp); !ok {
			return false, err
		} else if err != nil {
			return true, err
		}
	}

	return true, err
}

// RouteShutdown notifyes plugins of RouteShutdown routine
func (b *PluginBroker) RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	defer b.recover(ctx, rqs, rsp)

	for _, plugin := range b.Plugins() {
		if ok, err = plugin.RouteShutdown(ctx, rqs, rsp); !ok {
			return false, err
		} else if err != nil {
			return true, err
		}
	}

	return true, err
}

// DispatchLoopStartup notifyes plugins of DispatchLoopStartup routine
func (b *PluginBroker) DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	defer b.recover(ctx, rqs, rsp)

	for _, plugin := range b.Plugins() {
		if ok, err = plugin.DispatchLoopStartup(ctx, rqs, rsp); !ok {
			return false, err
		} else if err != nil {
			return true, err
		}
	}

	return true, err
}

// PreDispatch notifyes plugins of PreDispatch routine
func (b *PluginBroker) PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	defer b.recover(ctx, rqs, rsp)

	for _, plugin := range b.Plugins() {
		if ok, err = plugin.PreDispatch(ctx, rqs, rsp); !ok {
			return false, err
		} else if err != nil {
			return true, err
		}
	}

	return true, err
}

// PostDispatch notifyes plugins of PostDispatch routine
func (b *PluginBroker) PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	defer b.recover(ctx, rqs, rsp)

	for _, plugin := range b.Plugins() {
		if ok, err = plugin.PostDispatch(ctx, rqs, rsp); !ok {
			return false, err
		} else if err != nil {
			return true, err
		}
	}

	return true, err
}

// DispatchLoopShutdown notifyes plugins of DispatchLoopShutdown routine
func (b *PluginBroker) DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	defer b.recover(ctx, rqs, rsp)

	for _, plugin := range b.Plugins() {
		if ok, err = plugin.DispatchLoopShutdown(ctx, rqs, rsp); !ok {
			return false, err
		} else if err != nil {
			return true, err
		}
	}

	return true, err
}

func (b *PluginBroker) recover(ctx context.Context, rqs request.Interface, rsp response.Interface) (ok bool, err error) {
	if r := recover(); r != nil {
		switch err := r.(type) {
		case error:
			return false, errors.Wrap(err, "Unxpected error equired")

		default:
			return false, errors.Errorf("Unxpected error equired: %v", err)
		}
	}

	return true, nil
}

// NewPluginBroker creates a new plugins broker
func NewPluginBroker() (*PluginBroker, error) {
	return &PluginBroker{
		plugins: stack.NewPrioritised(),
	}, nil
}
