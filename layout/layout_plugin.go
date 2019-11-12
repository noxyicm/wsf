package layout

import (
	"fmt"
	"wsf/context"
	"wsf/controller/action/helper"
	"wsf/controller/plugin"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
)

// TYPELayoutPlugin is a plugin id
const TYPELayoutPlugin = "layout"

// NewLayoutPlugin  creates a new controller plugin layout
func NewLayoutPlugin() (plugin.Interface, error) {
	return &Plugin{
		name: TYPELayoutPlugin,
	}, nil
}

// Plugin is a controller plugin layout
type Plugin struct {
	name         string
	layout       Interface
	actionHelper helper.Interface
}

// Name returns plugin name
func (p *Plugin) Name() string {
	return p.name
}

// SetLayout sets a reference for layout instance
func (p *Plugin) SetLayout(l Interface) error {
	p.layout = l
	return nil
}

// GetLayout returns a reference for layout instance
func (p *Plugin) GetLayout() Interface {
	return p.layout
}

// SetLayoutActionHelper sets a reference for layout action helper instance
func (p *Plugin) SetLayoutActionHelper(h helper.Interface) error {
	p.actionHelper = h
	return nil
}

// GetLayoutActionHelper returns a reference for layout action helper instance
func (p *Plugin) GetLayoutActionHelper() helper.Interface {
	return p.actionHelper
}

// RouteStartup routine
func (p *Plugin) RouteStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// RouteShutdown routine
func (p *Plugin) RouteShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// DispatchLoopStartup routine
func (p *Plugin) DispatchLoopStartup(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}

// PreDispatch routine
func (p *Plugin) PreDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	ctx.SetValue(context.LayoutEnabledKey, true)
	return true, nil
}

// PostDispatch routine
func (p *Plugin) PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
	fmt.Println("layout_plugin.PostDispatch")
	l := p.GetLayout()
	if l == nil {
		return false, errors.New("[Layout] Layout object for plugin is not set")
	}

	// Return early if forward detected
	if !rqs.IsDispatched() || rsp.IsRedirect() {
		return true, nil
	}

	// Return early if layout has been disabled
	if enabled, _ := ctx.Value(context.LayoutEnabledKey).(bool); enabled && l.IsEnabled() {
		return true, nil
	}

	contentKey := l.GetContentKey()
	content := rsp.GetBody()

	if v, ok := content["default"]; ok {
		content[contentKey] = v
	}

	if contentKey == "default" {
		delete(content, "default")
	}

	for k, v := range content {
		ctx.SetDataValue(k, v)
	}

	fullContent := make([]byte, 0)
	var err error
	if name, ok := ctx.Value(context.LayoutKey).(string); ok {
		fullContent, err = l.Render(ctx, name)
	} else {
		err = errors.New("[Layout] Bad layout name")
	}

	if err != nil {
		rsp.SetData("layoutFullContent", fullContent)
		rsp.SetData("layoutContent", l.Get(contentKey))
		return false, err
	}

	rsp.SetBody(fullContent)
	return true, nil
}

// DispatchLoopShutdown routine
func (p *Plugin) DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}
