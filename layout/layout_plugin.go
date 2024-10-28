package layout

import (
	"wsf/context"
	"wsf/controller"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
)

// TYPEPluginLayout is a plugin id
const TYPEPluginLayout = "layout"

// NewLayoutPlugin  creates a new controller plugin layout
func NewLayoutPlugin() (controller.PluginInterface, error) {
	return &Plugin{
		name: "Layout",
	}, nil
}

// Plugin is a controller plugin layout
type Plugin struct {
	name         string
	layout       Interface
	actionHelper controller.HelperInterface
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

// SetLayoutHelper sets a reference for layout action helper instance
func (p *Plugin) SetLayoutHelper(h controller.HelperInterface) error {
	p.actionHelper = h
	return nil
}

// GetLayoutHelper returns a reference for layout action helper instance
func (p *Plugin) GetLayoutHelper() controller.HelperInterface {
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
	ctx.SetParam(context.LayoutEnabledKey, true)
	return true, nil
}

// PostDispatch routine
func (p *Plugin) PostDispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	l := p.GetLayout()
	if l == nil {
		return true, errors.New("[Layout] Layout object for plugin is not set")
	}

	// Return early if forward detected
	if !rqs.IsDispatched() || rsp.IsRedirect() {
		return true, nil
	}

	// Return early if layout has been disabled
	if !ctx.ParamBool(context.LayoutEnabledKey) || !l.IsEnabled() {
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
		ctx.SetParam(k, v)
	}

	fullContent := make([]byte, 0)
	var err error
	name := ctx.ParamString(context.LayoutKey)
	if name != "" {
		fullContent, err = l.Render(ctx, name)
	} else {
		err = errors.Errorf("[Layout] Bad layout name '%s'", name)
	}

	if err != nil {
		rsp.SetData("layoutFullContent", fullContent)
		rsp.SetData("layoutContent", l.Get(contentKey))
		return true, err
	}

	rsp.SetBody(fullContent)
	return true, nil
}

// DispatchLoopShutdown routine
func (p *Plugin) DispatchLoopShutdown(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	return true, nil
}
