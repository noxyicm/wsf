package controller

import (
	"net/http"
	"reflect"

	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view"
)

// Public variables
var (
	Error401 = errors.NewHTTP(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	Error402 = errors.NewHTTP(http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
	Error403 = errors.NewHTTP(http.StatusText(http.StatusForbidden), http.StatusForbidden)
	Error404 = errors.NewHTTP(http.StatusText(http.StatusNotFound), http.StatusNotFound)
	Error405 = errors.NewHTTP(http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
)

// ActionControllerInterface interface
type ActionControllerInterface interface {
	Init(ctx context.Context) error
	SetLogger(l *log.Log) error
	Logger() *log.Log
	SetHelperBroker(broker *HelperBroker) error
	HelperBroker() *HelperBroker
	HasHelper(name string) bool
	Helper(name string) HelperInterface
	Dispatch(ctx context.Context, ctrl ActionControllerInterface, m reflect.Method) error
	Render(ctx context.Context) error
	SetView(v view.Interface) error
	SetViewSuffix(suffix string) error
	View() view.Interface
	Invoke(ctx context.Context, ctrl ActionControllerInterface, m reflect.Method) error
	GetResource(name string) interface{}
}

// ActionControllerBase is a controller
type ActionControllerBase struct {
	logger       *log.Log
	InvokeParams map[string]interface{}
	Rqs          request.Interface
	Rsp          response.Interface
	Hlpr         *HelperBroker
	ViewSuffix   string
	Vw           view.Interface
}

// SetLogger attaches log writer
func (c *ActionControllerBase) SetLogger(l *log.Log) error {
	c.logger = l
	return nil
}

// Logger retreives attached log writer
func (c *ActionControllerBase) Logger() *log.Log {
	return c.logger
}

// Dispatch processes action call
func (c *ActionControllerBase) Dispatch(ctx context.Context, ctrl ActionControllerInterface, m reflect.Method) error {
	// Notify helpers of action preDispatch state
	err := c.Hlpr.NotifyPreDispatch(ctx)
	if err != nil {
		return err
	} else if err := ctx.Error(); err != nil {
		return err
	}

	err = c.PreDispatch(ctx)
	if err != nil {
		return err
	}

	if ctx.Request().IsDispatched() {
		// If pre-dispatch hooks introduced a redirect then stop dispatch
		if !ctx.Response().IsRedirect() {
			err = c.Invoke(ctx, ctrl, m)
			if err != nil {
				return err
			} else if err := ctx.Error(); err != nil {
				return err
			}
		}

		err = c.PostDispatch(ctx)
		if err != nil {
			return err
		}
	}

	// whats actually important here is that this action controller is
	// shutting down, regardless of dispatching; notify the helpers of this
	// state
	err = c.Hlpr.NotifyPostDispatch(ctx)
	if err != nil {
		return err
	} else if err := ctx.Error(); err != nil {
		return err
	}

	return nil
}

// SetHelperBroker sets helper broker
func (c *ActionControllerBase) SetHelperBroker(broker *HelperBroker) (err error) {
	if broker != nil {
		c.Hlpr = broker
	} else {
		c.Hlpr, err = NewHelperBroker()
		if err != nil {
			return err
		}
	}

	return nil
}

// HelperBroker returns action controller helper broker
func (c *ActionControllerBase) HelperBroker() *HelperBroker {
	return c.Hlpr
}

// HasHelper returns true if Action Halper is registered
func (c *ActionControllerBase) HasHelper(name string) bool {
	return c.Hlpr.HasHelper(name)
}

// Helper returns Action Halper
func (c *ActionControllerBase) Helper(name string) HelperInterface {
	h, _ := c.Hlpr.GetHelper(name)
	return h
}

// SetLayout sets a layout
func (c *ActionControllerBase) SetLayout(ctx context.Context, name string) error {
	ctx.SetParam(context.LayoutKey, name)
	return nil
}

// SetParams sets parameters to pass to handlers
func (c *ActionControllerBase) SetParams(params map[string]interface{}) error {
	c.InvokeParams = utils.MapSMerge(c.InvokeParams, params)
	return nil
}

// SetParam add or modify a parameter to use when instantiating a handler
func (c *ActionControllerBase) SetParam(name string, value interface{}) error {
	c.InvokeParams[name] = value
	return nil
}

// Param retrieve a single parameter from the parameter stack
func (c *ActionControllerBase) Param(name string) interface{} {
	if v, ok := c.InvokeParams[name]; ok {
		return v
	}

	return nil
}

// ParamString retrieve a single parameter from the parameter stack as string
func (c *ActionControllerBase) ParamString(name string) string {
	if v, ok := c.InvokeParams[name]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// ParamBool retrieve a single parameter from the parameter stack as boolean
func (c *ActionControllerBase) ParamBool(name string) bool {
	if v, ok := c.InvokeParams[name]; ok {
		if v, ok := v.(bool); ok {
			return v
		}

		return false
	}

	return false
}

// Params retrieve handler parameters
func (c *ActionControllerBase) Params() map[string]interface{} {
	return c.InvokeParams
}

// ClearParam clears the specified parameter
func (c *ActionControllerBase) ClearParam(name string) bool {
	if _, ok := c.InvokeParams[name]; ok {
		delete(c.InvokeParams, name)
		return true
	}

	return false
}

// ClearParams clears the parameter stack
func (c *ActionControllerBase) ClearParams() bool {
	c.InvokeParams = make(map[string]interface{})
	return true
}

// Init initializes controller
func (c *ActionControllerBase) Init(ctx context.Context) error {
	return nil
}

// Render renders response
func (c *ActionControllerBase) Render(ctx context.Context) error {
	return nil
}

// SetView sets action controller view object
func (c *ActionControllerBase) SetView(v view.Interface) error {
	c.Vw = v
	return nil
}

// SetViewSuffix sets action controller view suffix
func (c *ActionControllerBase) SetViewSuffix(suffix string) error {
	c.ViewSuffix = suffix
	return nil
}

// View returns action controller view object
func (c *ActionControllerBase) View() view.Interface {
	return c.Vw
}

// Invoke calls an action
func (c *ActionControllerBase) Invoke(ctx context.Context, ctrl ActionControllerInterface, m reflect.Method) error {
	defer c.invokeRecover(ctx)

	if err := c.verifySignature(m); err != nil {
		return err
	}

	values, err := c.resolveValues(m, ctrl, ctx)
	if err != nil {
		return err
	}

	out := m.Func.Call(values)
	if out[0].IsNil() {
		return nil
	}

	return out[0].Interface().(error)
}

// ViewScript returns path to view script
func (c *ActionControllerBase) ViewScript(ctx context.Context, action string, noController bool) (string, error) {
	if !c.ParamBool("noViewRenderer") && c.Hlpr.HasHelper("ViewRenderer") {
		viewRenderer, err := c.Hlpr.GetHelper("viewRenderer")
		if err != nil {
			return "", err
		}

		if noController {
			viewRenderer.(*ViewRenderer).SetNoController(noController)
		}

		if action != "" {
			return viewRenderer.(*ViewRenderer).ViewScript(map[string]string{
				"module":     ctx.Request().ModuleName(),
				"controller": ctx.Request().ControllerName(),
				"action":     action,
			})
		}

		return viewRenderer.(*ViewRenderer).ViewScript(map[string]string{
			"module":     ctx.Request().ModuleName(),
			"controller": ctx.Request().ControllerName(),
			"action":     ctx.Request().ActionName(),
		})
	}

	rqs := ctx.Request()
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
func (c *ActionControllerBase) GetResource(name string) interface{} {
	return registry.GetResource(name)
}

// PreDispatch fires before action invocation
func (c *ActionControllerBase) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch fires after action invocation
func (c *ActionControllerBase) PostDispatch(ctx context.Context) error {
	return nil
}

// verifySignature checks if action method has valid signature
func (c *ActionControllerBase) verifySignature(m reflect.Method) error {
	if m.Type.NumIn() < 1 {
		return errors.Errorf("Action ( %s ) must have atleast 1 value", m.Name)
	}

	if !m.Type.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return errors.Errorf("Action ( %s ) first argument must implement context.Context interface", m.Name)
	}

	if m.Type.NumOut() != 1 {
		return errors.Errorf("Action ( %s ) must have exact 1 return value", m.Name)
	}

	if !m.Type.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.Errorf("Action ( %s ) return value must be error type", m.Name)
	}

	return nil
}

// resolveValues returns slice of call arguments for service Init method
func (c *ActionControllerBase) resolveValues(m reflect.Method, args ...interface{}) (values []reflect.Value, err error) {
	for i := 0; i < m.Type.NumIn(); i++ {
		v := m.Type.In(i)

		//switch {
		//case v.ConvertibleTo(reflect.TypeOf(ctrl)):
		//	values = append(values, reflect.ValueOf(ctrl))

		//case v.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()):
		//	values = append(values, reflect.ValueOf(cfg))

		//default:
		if len(args) > i {
			value, err := c.resolveValue(v, args[i])
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		} else {
			values = append(values, reflect.Value{})
		}
		//}
	}

	return
}

func (c *ActionControllerBase) resolveValue(v reflect.Type, arg interface{}) (reflect.Value, error) {
	value := reflect.Value{}
	if v.ConvertibleTo(reflect.TypeOf(arg)) {
		value = reflect.ValueOf(arg)
	} else if v.Kind() == reflect.Interface && reflect.TypeOf(arg).Implements(v) {
		value = reflect.ValueOf(arg)
	}

	if !value.IsValid() {
		value = reflect.New(v).Elem()
	}

	return value, nil
}

/*func (c *ActionControllerBase) initView() (vi view.Interface, err error) {
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
}*/

func (c *ActionControllerBase) invokeRecover(ctx context.Context) {
	if r := recover(); r != nil {
		switch er := r.(type) {
		case error:
			ctx.AddError(errors.Wrap(er, "Unxpected error equired"))

		default:
			ctx.AddError(errors.Errorf("Unxpected error equired: %v", er))
		}
	}
}

// NewActionControllerBase creates an instance of action controller
func NewActionControllerBase() (c *ActionControllerBase, err error) {
	c = &ActionControllerBase{}

	untypedLog := registry.GetResource("syslog")
	if untypedLog == nil {
		return nil, errors.New("Log resource is required")
	}
	c.SetLogger(untypedLog.(*log.Log))

	return c, nil
}
