package helper

import (
	"regexp"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/filter"
	"wsf/filter/word"
	"wsf/registry"
	"wsf/session"
	"wsf/utils"
	"wsf/view"
)

const (
	// TYPEViewRenderer represents ViewRenderer action helper
	TYPEViewRenderer = "viewRenderer"
)

func init() {
	Register(TYPEViewRenderer, NewViewRenderer)
}

// ViewRenderer is a action helper that handles render
type ViewRenderer struct {
	name                           string
	actionController               ControllerInterface
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
	inflector                      *filter.Inflector
	View                           view.Interface
}

// Name returns helper name
func (vr *ViewRenderer) Name() string {
	return vr.name
}

// SetController sets action controller
func (vr *ViewRenderer) SetController(ctrl ControllerInterface) error {
	vr.actionController = ctrl
	return nil
}

// Controller returns action controller
func (vr *ViewRenderer) Controller() ControllerInterface {
	return vr.actionController
}

// Init the helper
func (vr *ViewRenderer) Init(options map[string]interface{}) error {
	return vr.initView("", "", options)
}

// PreDispatch do dispatch preparations
func (vr *ViewRenderer) PreDispatch() error {
	return nil
}

// PostDispatch do dispatch aftermath
func (vr *ViewRenderer) PostDispatch() error {
	if vr.shouldRender() {
		return vr.Render(vr.Request().ActionName(), "default", false)
	}

	return nil
}

// SetInflector sets the inflector
func (vr *ViewRenderer) SetInflector(inflector *filter.Inflector) error {
	vr.inflector = inflector
	return nil
}

// GetInflector returns filter inflector
func (vr *ViewRenderer) GetInflector() *filter.Inflector {
	if vr.inflector == nil {
		inflector, err := filter.NewInflector()
		if err != nil {
			return nil
		}

		vr.inflector = inflector.(*filter.Inflector)
		vr.inflector.SetStaticRule("moduleDir", vr.moduleDir) // moduleDir must be specified before the less specific 'module'

		uts, err := word.NewUnderscoreToSeparator("/")
		if err != nil {
			return nil
		}

		rrc, err := filter.NewRegexpReplace(`\.`, "-")
		if err != nil {
			return nil
		}

		rra, err := filter.NewRegexpReplace(regexp.QuoteMeta(`[^a-z0-9#]+`), "-")
		if err != nil {
			return nil
		}

		vr.inflector.AddRules(map[string]interface{}{
			":module":     []interface{}{"Word_CamelCaseToDash", "StringToLower"},
			":controller": []interface{}{"Word_CamelCaseToDash", uts, "StringToLower", rrc},
			":action":     []interface{}{"Word_CamelCaseToDash", rra, "StringToLower"},
		})
		vr.inflector.SetStaticRule("suffix", vr.viewSuffix)
		vr.inflector.SetTarget(vr.inflectorTarget)
	}

	return vr.inflector
}

// SetNoRender sets view renderer state
func (vr *ViewRenderer) SetNoRender(state bool) error {
	vr.noRender = state
	return nil
}

// SetNoController sets view renderer controller state
func (vr *ViewRenderer) SetNoController(state bool) error {
	vr.noController = state
	return nil
}

// Request returns request object
func (vr *ViewRenderer) Request() request.Interface {
	return vr.Controller().Request()
}

// Response return response object
func (vr *ViewRenderer) Response() response.Interface {
	return vr.Controller().Response()
}

// Session return session object
func (vr *ViewRenderer) Session() session.Interface {
	return vr.Controller().Session()
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

// SetScriptAction sets renderer script action
func (vr *ViewRenderer) SetScriptAction(action string) error {
	vr.scriptAction = action
	return nil
}

// GetScriptAction return renderer script action
func (vr *ViewRenderer) GetScriptAction() string {
	return vr.scriptAction
}

// GetViewScript returns renderer view script
func (vr *ViewRenderer) GetViewScript(action string, params map[string]string) (string, error) {
	request := vr.Request()
	if _, ok := params["action"]; action == "" && !ok {
		action = vr.GetScriptAction()
		if action == "" {
			action = request.ActionName()
		}

		params["action"] = action
	} else if action != "" {
		params["action"] = action
	}

	replacePatternStart, _ := regexp.Compile(`[^a-z0-9]+$`)
	replacePatternEnd, _ := regexp.Compile(`^[^a-z0-9]+`)
	params["action"] = replacePatternStart.ReplaceAllString(params["action"], "")
	params["action"] = replacePatternEnd.ReplaceAllString(params["action"], "")

	_ = vr.GetInflector()
	if vr.noController || vr.neverController {
		vr.setInflectorTarget(vr.viewScriptPathNoControllerSpec)
	} else {
		vr.setInflectorTarget(vr.viewScriptPathSpec)
	}

	return vr.translateSpec(params)
}

// RenderScript renders script
func (vr *ViewRenderer) RenderScript(script string, name string) error {
	if name == "" {
		name = vr.GetResponseSegment()
	}

	if vr.View == nil {
		return errors.New("View is not initialized")
	}

	vr.Controller().Context().SetValue("renderedscript", script)
	rendered, err := vr.View.Render(vr.Controller().Context(), script)
	if err != nil {
		return err
	}

	vr.Response().AppendBody(rendered, name)
	vr.SetNoRender(true)
	return nil
}

// Render the script for action
func (vr *ViewRenderer) Render(action string, name string, noController bool) error {
	//vr.setRender(action, name, noController)
	path, err := vr.GetViewScript(action, map[string]string{})
	if err != nil {
		return errors.Wrap(err, "[ViewRenderer] Render error")
	}

	return vr.RenderScript(path, name)
}

// Should the ViewRenderer render a view script?
func (vr *ViewRenderer) shouldRender() bool {
	return (vr.actionController != nil && !vr.actionController.ParamBool("noViewRenderer") && !vr.neverRender && !vr.noRender && vr.Request().IsDispatched() && !vr.Response().IsRedirect())
}

func (vr *ViewRenderer) setRender(action, name string, noController bool) *ViewRenderer {
	if action != "" {
		vr.SetScriptAction(action)
	}

	if name != "" {
		vr.SetResponseSegment(name)
	}

	vr.SetNoController(noController)

	return vr
}

func (vr *ViewRenderer) setInflectorTarget(target string) error {
	vr.inflectorTarget = target
	if vr.inflector != nil {
		return vr.inflector.SetTarget(target)
	}

	return nil
}

func (vr *ViewRenderer) setModuleDir(dir string) error {
	vr.moduleDir = dir
	return nil
}

func (vr *ViewRenderer) translateSpec(vars map[string]string) (string, error) {
	inflector := vr.GetInflector()
	rqs := vr.Request()

	// Format controller name
	filter, err := word.NewCamelCaseToDash()
	if err != nil {
		return "", err
	}

	controller, err := filter.Filter(rqs.ControllerName())
	if err != nil {
		return "", err
	}

	params := map[string]string{
		"module":     rqs.ModuleName(),
		"controller": controller.(string),
		"action":     rqs.ActionName(),
	}

	for key, value := range vars {
		switch key {
		case "module", "controller", "action", "moduleDir", "suffix":
			params[key] = value
		}
	}

	var origSuffix, origModuleDir string
	if params["suffix"] != "" {
		origSuffix = vr.viewSuffix
		vr.viewSuffix = params["suffix"]
	}

	if params["moduleDir"] != "" {
		origModuleDir = vr.moduleDir
		vr.moduleDir = params["moduleDir"]
	}

	filtered, err := inflector.Filter(params)
	if err != nil {
		return "", err
	}

	if origSuffix != "" {
		vr.viewSuffix = origSuffix
	}

	if origModuleDir != "" {
		vr.moduleDir = origModuleDir
	}

	return filtered.(string), nil
}

func (vr *ViewRenderer) initView(path string, prefix string, options map[string]interface{}) (err error) {
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

	// Get base view path
	if path == "" {
		path, err = vr.getBasePath()
		if err != nil {
			return err
		}

		if path == "" {
			return errors.Errorf("ViewRenderer initialization failed: retrieved view base path is empty")
		}
	}

	vr.moduleDir = path

	// Register view with action controller (unless already registered)
	if vr.actionController != nil && vr.actionController.View() == nil {
		vr.actionController.SetView(vr.View)
		vr.actionController.SetViewSuffix(vr.viewSuffix)
	}

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

func (vr *ViewRenderer) getBasePath() (string, error) {
	if vr.actionController == nil {
		return "views", nil
	}

	inflector := vr.GetInflector()
	vr.setInflectorTarget(vr.viewBasePathSpec)

	rqs := vr.Request()
	parts := map[string]string{
		"module":     rqs.ModuleName(),
		"controller": rqs.ControllerName(),
		"action":     rqs.ActionName(),
	}

	path, err := inflector.Filter(parts)
	if err != nil {
		return "", err
	}

	return path.(string), nil
}

func (vr *ViewRenderer) generateDefaultPrefix() string {
	defaultPrefix := "View"
	if vr.actionController == nil {
		return defaultPrefix
	}

	return defaultPrefix
}

// NewViewRenderer creates new ViewRenderer action helper
func NewViewRenderer() (Interface, error) {
	return &ViewRenderer{
		name: "viewRenderer",
	}, nil
}
