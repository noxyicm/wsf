package controller

import (
	"regexp"

	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/filter"
	"github.com/noxyicm/wsf/filter/word"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view"
)

const (
	// TYPEHelperViewRenderer represents ViewRenderer action helper
	TYPEHelperViewRenderer = "viewRenderer"
)

var (
	replacePatternStart = regexp.MustCompile(`[^a-z0-9]+$`)
	replacePatternEnd   = regexp.MustCompile(`^[^a-z0-9]+`)
)

func init() {
	RegisterHelper(TYPEHelperViewRenderer, NewViewRendererHelper)
}

// ViewRenderer is a action helper that handles render
type ViewRenderer struct {
	name                           string
	moduleDir                      string
	inflectorTarget                string
	viewScriptPathNoControllerSpec string
	viewScriptPathSpec             string
	viewBasePathSpec               string
	neverController                bool
	neverRender                    bool
	noController                   bool
	noRender                       bool
	responseSegment                string
	scriptAction                   string
	viewSuffix                     string
	View                           view.Interface
}

// Name returns helper name
func (vr *ViewRenderer) Name() string {
	return vr.name
}

// Init the helper
func (vr *ViewRenderer) Init(options map[string]interface{}) error {
	return vr.initView(options)
}

// PreDispatch do dispatch preparations
func (vr *ViewRenderer) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (vr *ViewRenderer) PostDispatch(ctx context.Context) error {
	if vr.shouldRender(ctx) {
		if ctx.ParamBool(context.LayoutEnabledKey) {
			if layout := ctx.ParamString(context.LayoutKey); layout != "" {
				return vr.Render(ctx, layout, false)
			}
		}

		return vr.Render(ctx, vr.View.GetOptions().LayoutContentKey, false)
	}

	return nil
}

// SetNoRender sets render state
func (vr *ViewRenderer) SetNoRender(state bool) error {
	vr.noRender = state
	return nil
}

// SetNoController sets controller state
func (vr *ViewRenderer) SetNoController(state bool) error {
	vr.noController = state
	return nil
}

// SetResponseSegment sets rendering segment
func (vr *ViewRenderer) SetResponseSegment(name string) error {
	vr.responseSegment = name
	return nil
}

// GetResponseSegment returns rendering segment
func (vr *ViewRenderer) GetResponseSegment() string {
	return vr.responseSegment
}

// ViewScript returns renderer view script
func (vr *ViewRenderer) ViewScript(params map[string]string) (string, error) {
	if _, ok := params["action"]; !ok {
		return "", errors.New("Action must be specified")
	}

	if _, ok := params["controller"]; !ok {
		return "", errors.New("Controller must be specified")
	}

	if _, ok := params["module"]; !ok {
		return "", errors.New("Module must be specified")
	}

	params["action"] = replacePatternStart.ReplaceAllString(params["action"], "")
	params["action"] = replacePatternEnd.ReplaceAllString(params["action"], "")

	path, err := vr.getBasePath(params)
	if err != nil {
		return "", errors.Wrap(err, "Unable to get script path")
	}

	inflector, err := filter.NewInflector()
	if err != nil {
		return "", errors.Wrap(err, "Unable to get script path")
	}
	inflector.SetStaticRule("moduleDir", path) // moduleDir must be specified before the less specific 'module'

	uts, err := word.NewUnderscoreToSeparator("/")
	if err != nil {
		return "", errors.Wrap(err, "Unable to get script path")
	}

	rrc, err := filter.NewRegexpReplace(`\.`, "-")
	if err != nil {
		return "", errors.Wrap(err, "Unable to get script path")
	}

	rra, err := filter.NewRegexpReplace(regexp.QuoteMeta(`[^a-z0-9#]+`), "-")
	if err != nil {
		return "", errors.Wrap(err, "Unable to get script path")
	}

	inflector.AddRules(map[string]interface{}{
		":module":     []interface{}{"Word_CamelCaseToDash", "StringToLower"},
		":controller": []interface{}{"Word_CamelCaseToDash", uts, "StringToLower", rrc},
		":action":     []interface{}{"Word_CamelCaseToDash", rra, "StringToLower"},
	})
	inflector.SetStaticRule("suffix", vr.viewSuffix)

	var inflectorTarget string
	if vr.noController || vr.neverController {
		inflectorTarget = vr.viewScriptPathNoControllerSpec
	} else {
		inflectorTarget = vr.viewScriptPathSpec
	}

	inflector.SetTarget(inflectorTarget)

	filtered, err := inflector.Filter(params)
	if err != nil {
		return "", errors.Wrap(err, "Unable to get script path")
	}

	return filtered.(string), nil
}

// RenderScript renders script
func (vr *ViewRenderer) RenderScript(ctx context.Context, script string, name string) error {
	if name == "" {
		name = vr.GetResponseSegment()
	}

	if vr.View == nil {
		return errors.New("View is not initialized")
	}

	rendered, err := vr.View.Render(ctx, script, name)
	if err != nil {
		return err
	}

	ctx.Response().AppendBody(rendered, vr.GetResponseSegment())
	ctx.SetParam("noRender", true)
	return nil
}

// Render the script for action
func (vr *ViewRenderer) Render(ctx context.Context, name string, noController bool) error {
	path, err := vr.ViewScript(map[string]string{
		"module":     ctx.Request().ModuleName(),
		"controller": ctx.Request().ControllerName(),
		"action":     ctx.Request().ActionName(),
	})
	if err != nil {
		return errors.Wrap(err, "[ViewRenderer] Render error")
	}

	return vr.RenderScript(ctx, path, name)
}

// ViewSuffix retrives view suffix
func (vr *ViewRenderer) ViewSuffix() string {
	return vr.viewSuffix
}

// SetViewSuffix sets view suffix
func (vr *ViewRenderer) SetViewSuffix(suffix string) error {
	vr.viewSuffix = suffix
	return nil
}

// Should the ViewRenderer render a view script?
func (vr *ViewRenderer) shouldRender(ctx context.Context) bool {
	if enabled, _ := ctx.Value(context.NoRenderKey).(bool); enabled {
		return false
	}

	return !ctx.ParamBool("noRender") && !ctx.ParamBool("noViewRenderer") && !vr.neverRender && !vr.noRender && ctx.Request().IsDispatched() && !ctx.Response().IsRedirect()
}

func (vr *ViewRenderer) initView(options map[string]interface{}) (err error) {
	if vr.View == nil {
		v := registry.GetResource("view")
		if v == nil {
			return errors.New("View resource is not initialized")
		}

		vr.View = v.(view.Interface)
	}

	defaultOptions := map[string]interface{}{
		"neverRender":                    false,
		"neverController":                false,
		"noController":                   false,
		"noRender":                       false,
		"scriptAction":                   "",
		"responseSegment":                "",
		"viewBasePathSpec":               vr.View.GetBasePath(),
		"viewScriptPathSpec":             vr.View.GetScriptPath(),
		"viewScriptPathNoControllerSpec": vr.View.GetScriptPathNoController(),
		"viewSuffix":                     vr.View.GetSuffix(),
	}

	if options == nil {
		options = defaultOptions
	} else {
		options = utils.MapSMerge(defaultOptions, options)
	}

	// Set options first; may be used to determine other initializations
	vr.setOptions(options)
	return nil
}

func (vr *ViewRenderer) setOptions(options map[string]interface{}) error {
	for key, value := range options {
		switch key {
		case "neverRender", "neverController", "noController", "noRender":
			param := false
			if v, ok := value.(bool); ok {
				param = v
			}

			switch key {
			case "neverRender":
				vr.neverRender = param

			case "neverController":
				vr.neverController = param

			case "noController":
				vr.noController = param

			case "noRender":
				vr.noRender = param
			}

		case "responseSegment", "scriptAction", "viewBasePathSpec", "viewScriptPathSpec", "viewScriptPathNoControllerSpec", "viewSuffix":
			param := ""
			if v, ok := value.(string); ok {
				param = v
			}

			switch key {
			case "responseSegment":
				vr.responseSegment = param

			case "scriptAction":
				vr.scriptAction = param

			case "viewBasePathSpec":
				vr.viewBasePathSpec = param

			case "viewScriptPathSpec":
				vr.viewScriptPathSpec = param

			case "viewScriptPathNoControllerSpec":
				vr.viewScriptPathNoControllerSpec = param

			case "viewSuffix":
				vr.viewSuffix = param
			}
		}
	}

	return nil
}

func (vr *ViewRenderer) getBasePath(params map[string]string) (string, error) {
	inflector, err := filter.NewInflector()
	if err != nil {
		return "", errors.Wrap(err, "Unable to get controller path for view templates")
	}

	uts, err := word.NewUnderscoreToSeparator("/")
	if err != nil {
		return "", errors.Wrap(err, "Unable to get controller path for view templates")
	}

	rrc, err := filter.NewRegexpReplace(`\.`, "-")
	if err != nil {
		return "", errors.Wrap(err, "Unable to get controller path for view templates")
	}

	inflector.AddRules(map[string]interface{}{
		":module":     []interface{}{"Word_CamelCaseToDash", "StringToLower"},
		":controller": []interface{}{"Word_CamelCaseToDash", uts, "StringToLower", rrc},
	})
	inflector.SetTarget(vr.viewBasePathSpec)

	controllerPath, err := inflector.Filter(params)
	if err != nil {
		return "", errors.Wrap(err, "Unable to add controller path for view templates")
	}

	return controllerPath.(string), nil
}

// NewViewRendererHelper creates new ViewRenderer action helper
func NewViewRendererHelper() (HelperInterface, error) {
	return &ViewRenderer{
		name: "ViewRenderer",
	}, nil
}
