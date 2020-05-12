package action

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"wsf/context"
	"wsf/controller/action/helper"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/registry"
	"wsf/session"
	"wsf/utils"
	"wsf/view"
)

// Public variables
var (
	Error401 = errors.NewHTTP(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	Error402 = errors.NewHTTP(http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
	Error403 = errors.NewHTTP(http.StatusText(http.StatusForbidden), http.StatusForbidden)
	Error404 = errors.NewHTTP(http.StatusText(http.StatusNotFound), http.StatusNotFound)
	Error405 = errors.NewHTTP(http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
)

// Interface controller interface
type Interface interface {
	Init() error
	SetRequest(req request.Interface)
	Request() request.Interface
	SetResponse(rsp response.Interface)
	Response() response.Interface
	SetSession(s session.Interface)
	Session() session.Interface
	SetContext(ctx context.Context)
	Context() context.Context
	SetParams(params map[string]interface{}) error
	SetParam(name string, value interface{}) error
	Param(name string) interface{}
	ParamString(name string) string
	ParamBool(name string) bool
	Params() map[string]interface{}
	ClearParam(name string) bool
	ClearParams() bool
	SetHelperBroker() error
	HelperBroker() *HelperBroker
	HasHelper(name string) bool
	Helper(name string) helper.Interface
	Dispatch(ctrl interface{}, m reflect.Method) error
	Render() error
	SetView(v view.Interface) error
	SetViewSuffix(suffix string) error
	View() view.Interface
	Invoke(ctrl interface{}, m reflect.Method) error
	GetResource(name string) interface{}
}

// Controller controller
type Controller struct {
	InvokeParams map[string]interface{}
	Rqs          request.Interface
	Rsp          response.Interface
	Ctx          context.Context
	Sess         session.Interface
	Hlpr         *HelperBroker
	ViewSuffix   string
	Vw           view.Interface
}

// Dispatch processes action call
func (c *Controller) Dispatch(ctrl interface{}, m reflect.Method) error {
	// Notify helpers of action preDispatch state
	err := c.Hlpr.NotifyPreDispatch()
	if err != nil {
		return err
	}

	err = c.PreDispatch()
	if err != nil {
		return err
	}

	if c.Rqs.IsDispatched() {
		// If pre-dispatch hooks introduced a redirect then stop dispatch
		if !c.Rsp.IsRedirect() {
			err = c.Invoke(ctrl, m)
			if err != nil {
				return err
			}
		}

		err = c.PostDispatch()
		if err != nil {
			return err
		}
	}

	// whats actually important here is that this action controller is
	// shutting down, regardless of dispatching; notify the helpers of this
	// state
	err = c.Hlpr.NotifyPostDispatch()
	if err != nil {
		return err
	}

	return nil
}

// SetHelperBroker sets helper broker
func (c *Controller) SetHelperBroker() (err error) {
	if broker != nil {
		c.Hlpr = broker
	} else {
		c.Hlpr, err = NewHelperBroker()
		if err != nil {
			return err
		}
	}

	c.Hlpr.SetController(c)
	return nil
}

// HelperBroker returns action controller helper broker
func (c *Controller) HelperBroker() *HelperBroker {
	return c.Hlpr
}

// HasHelper returns true if Action Halper is registered
func (c *Controller) HasHelper(name string) bool {
	return c.Hlpr.HasHelper(name)
}

// Helper returns Action Halper
func (c *Controller) Helper(name string) helper.Interface {
	h, _ := c.Hlpr.GetHelper(name)
	return h
}

// SetRequest sets request
func (c *Controller) SetRequest(req request.Interface) {
	c.Rqs = req
}

// Request returns request
func (c *Controller) Request() request.Interface {
	return c.Rqs
}

// SetResponse sets response
func (c *Controller) SetResponse(rsp response.Interface) {
	c.Rsp = rsp
}

// Response returns response
func (c *Controller) Response() response.Interface {
	return c.Rsp
}

// SetSession sets session
func (c *Controller) SetSession(s session.Interface) {
	c.Sess = s
}

// Session returns session
func (c *Controller) Session() session.Interface {
	return c.Sess
}

// SetContext sets context
func (c *Controller) SetContext(ctx context.Context) {
	c.Ctx = ctx
}

// Context returns context
func (c *Controller) Context() context.Context {
	return c.Ctx
}

// SetLayout sets a layout
func (c *Controller) SetLayout(name string) error {
	c.Ctx.SetValue(context.LayoutKey, name)
	return nil
}

// DisableLayout disables the layout
func (c *Controller) DisableLayout() {
	c.Ctx.SetValue(context.LayoutEnabledKey, false)
}

// SetParams sets parameters to pass to handlers
func (c *Controller) SetParams(params map[string]interface{}) error {
	c.InvokeParams = utils.MapSMerge(c.InvokeParams, params)
	return nil
}

// SetParam add or modify a parameter to use when instantiating a handler
func (c *Controller) SetParam(name string, value interface{}) error {
	c.InvokeParams[name] = value
	return nil
}

// Param retrieve a single parameter from the parameter stack
func (c *Controller) Param(name string) interface{} {
	if v, ok := c.InvokeParams[name]; ok {
		return v
	}

	return nil
}

// ParamString retrieve a single parameter from the parameter stack as string
func (c *Controller) ParamString(name string) string {
	if v, ok := c.InvokeParams[name]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// ParamBool retrieve a single parameter from the parameter stack as boolean
func (c *Controller) ParamBool(name string) bool {
	if v, ok := c.InvokeParams[name]; ok {
		if v, ok := v.(bool); ok {
			return v
		}

		return false
	}

	return false
}

// Params retrieve handler parameters
func (c *Controller) Params() map[string]interface{} {
	return c.InvokeParams
}

// ClearParam clears the specified parameter
func (c *Controller) ClearParam(name string) bool {
	if _, ok := c.InvokeParams[name]; ok {
		delete(c.InvokeParams, name)
		return true
	}

	return false
}

// ClearParams clears the parameter stack
func (c *Controller) ClearParams() bool {
	c.InvokeParams = make(map[string]interface{})
	return true
}

// Init initializes controller
func (c *Controller) Init() error {
	return nil
}

// Render renders response
func (c *Controller) Render() error {
	return nil
}

// SetView sets action controller view object
func (c *Controller) SetView(v view.Interface) error {
	c.Vw = v
	return nil
}

// SetViewSuffix sets action controller view suffix
func (c *Controller) SetViewSuffix(suffix string) error {
	c.ViewSuffix = suffix
	return nil
}

// View returns action controller view object
func (c *Controller) View() view.Interface {
	return c.Vw
}

// Invoke calls an action
func (c *Controller) Invoke(ctrl interface{}, m reflect.Method) error {
	if err := c.verifySignature(m); err != nil {
		return err
	}

	values, err := c.resolveValues(ctrl, m)
	if err != nil {
		return err
	}

	out := m.Func.Call(values)
	if out[0].IsNil() {
		return nil
	}

	return out[0].Interface().(error)
}

// GetViewScript returns path to view script
func (c *Controller) GetViewScript(action string, noController bool) (string, error) {
	if !c.ParamBool("noViewRenderer") && c.Hlpr.HasHelper("viewRenderer") {
		viewRenderer, err := c.Hlpr.GetHelper("viewRenderer")
		if err != nil {
			return "", err
		}

		if noController {
			viewRenderer.(*helper.ViewRenderer).SetNoController(noController)
		}

		return viewRenderer.(*helper.ViewRenderer).GetViewScript(action, nil)
	}

	rqs := c.Request()
	if action == "" {
		action = rqs.ActionName()
	}

	if action == "" {
		return "", errors.New("Invalid action specifier for view render")
	}

	script := action + "." + c.ViewSuffix

	if !noController {
		controller := rqs.ControllerName()
		script = controller + "/" + script
	}

	return script, nil
}

// GetResource returns a registered resource from registry
func (c *Controller) GetResource(name string) interface{} {
	return registry.GetResource(name)
}

// PreDispatch fires before action invocation
func (c *Controller) PreDispatch() error {
	return nil
}

// PostDispatch fires after action invocation
func (c *Controller) PostDispatch() error {
	return nil
}

// verifySignature checks if action method has valid signature
func (c *Controller) verifySignature(m reflect.Method) error {
	if m.Type.NumOut() != 1 {
		return errors.Errorf("Action ( %s ) must have exact 1 return value", m.Name)
	}

	if !m.Type.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.Errorf("Action ( %s ) return value must be error type", m.Name)
	}

	return nil
}

// resolveValues returns slice of call arguments for service Init method
func (c *Controller) resolveValues(ctrl interface{}, m reflect.Method) (values []reflect.Value, err error) {
	for i := 0; i < m.Type.NumIn(); i++ {
		v := m.Type.In(i)

		switch {
		case v.ConvertibleTo(reflect.ValueOf(ctrl).Type()):
			values = append(values, reflect.ValueOf(ctrl))

		default:
			value, err := c.resolveValue(v)
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}
	}

	return
}

func (c *Controller) resolveValue(v reflect.Type) (reflect.Value, error) {
	value := reflect.Value{}
	return value, nil
}

func (c *Controller) initView() (vi view.Interface, err error) {
	if !c.ParamBool("noViewRenderer") && c.Hlpr.HasHelper("viewRenderer") {
		return nil, nil
	}

	if c.Vw != nil {
		return c.Vw, nil
	}

	rqs := c.Request()
	mdl := rqs.ModuleName()

	//os.PathSeparator
	baseDir := filepath.Dir(mdl) + "/views"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, errors.Errorf("Missing base view directory ( '%s' )", baseDir)
	}

	viewCfg := &view.Config{Type: "default", BaseDir: baseDir}
	c.Vw, err = view.NewDefaultView(viewCfg)
	if err != nil {
		return nil, err
	}

	registry.Set("view", c.Vw)
	return c.Vw, nil
}
