package controller

import (
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/layout"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view"
)

// TYPEHelperViewLayout is a plugin id
const TYPEHelperViewLayout = "viewLayout"

func init() {
	RegisterHelper(TYPEHelperViewLayout, NewViewLayoutHelper)
}

// Plugin is a controller plugin layout
type ViewLayout struct {
	name               string
	Layout             layout.Interface
	View               view.Interface
	viewScriptPathSpec string
	viewBasePathSpec   string
	viewSuffix         string
	noRender           bool
	neverRender        bool
}

// Name returns helper name
func (vl *ViewLayout) Name() string {
	return vl.name
}

// Init the helper
func (vl *ViewLayout) Init(options map[string]interface{}) error {
	return vl.initView(options)
}

// PreDispatch do dispatch preparations
func (vl *ViewLayout) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (vl *ViewLayout) PostDispatch(ctx context.Context) error {
	if vl.shouldRender(ctx) {
		if ctx.ParamBool(context.LayoutEnabledKey) {
			if layout := ctx.ParamString(context.LayoutKey); layout != "" {
				return vl.Render(ctx, layout)
			}

			return vl.Render(ctx, vl.Layout.GetOptions().Layout)
		}
	}

	return nil
}

func (vl *ViewLayout) Render(ctx context.Context, name string) error {
	if vl.Layout == nil {
		return errors.New("Layout is not set")
	}

	rsp := ctx.Response()
	body := rsp.GetBody()
	data := utils.MapSMerge(ctx.Data(), body)
	rendered, err := vl.Layout.Render(data, name)
	if err != nil {
		return errors.Wrap(err, "[ViewRenderer] Render error")
	}

	ctx.Response().SetBody(rendered)
	ctx.SetParam("noRender", true)
	return nil
}

func (vl *ViewLayout) initView(options map[string]interface{}) (err error) {
	if vl.View == nil {
		v := registry.GetResource("view")
		if v == nil {
			return errors.New("View resource is not initialized")
		}

		vl.View = v.(view.Interface)
	}

	if vl.Layout == nil {
		l := registry.GetResource("layout")
		if l == nil {
			return errors.New("Layout resource is not initialized")
		}

		vl.Layout = l.(layout.Interface)
	}

	defaultOptions := map[string]interface{}{
		"neverRender":        false,
		"noRender":           false,
		"viewBasePathSpec":   vl.Layout.GetViewBasePath(),
		"viewScriptPathSpec": vl.Layout.GetViewScriptPath(),
		"viewSuffix":         vl.Layout.GetViewSuffix(),
	}

	if options == nil {
		options = defaultOptions
	} else {
		options = utils.MapSMerge(defaultOptions, options)
	}

	// Set options first; may be used to determine other initializations
	vl.setOptions(options)
	return nil
}

func (vl *ViewLayout) setOptions(options map[string]interface{}) error {
	for key, value := range options {
		switch key {
		case "neverRender", "noRender":
			param := false
			if v, ok := value.(bool); ok {
				param = v
			}

			switch key {
			case "neverRender":
				vl.neverRender = param

			case "noRender":
				vl.noRender = param
			}

		case "viewBasePathSpec", "viewScriptPathSpec", "viewSuffix":
			param := ""
			if v, ok := value.(string); ok {
				param = v
			}

			switch key {
			case "viewBasePathSpec":
				vl.viewBasePathSpec = param

			case "viewScriptPathSpec":
				vl.viewScriptPathSpec = param

			case "viewSuffix":
				vl.viewSuffix = param
			}
		}
	}

	return nil
}

// Should the ViewLayout render a view script?
func (vl *ViewLayout) shouldRender(ctx context.Context) bool {
	if enabled, _ := ctx.Value(context.NoRenderKey).(bool); enabled {
		return false
	}

	return !ctx.ParamBool("noRender") && !ctx.ParamBool("noViewLayout") && !vl.neverRender && !vl.noRender && ctx.Request().IsDispatched() && !ctx.Response().IsRedirect()
}

// NewViewLayoutHelper creates new ViewLayout action helper
func NewViewLayoutHelper() (HelperInterface, error) {
	return &ViewLayout{
		name: "ViewLayout",
	}, nil
}
