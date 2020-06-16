package dispatcher

import (
	"reflect"
	"strings"
	"wsf/application/modules"
	"wsf/context"
	"wsf/controller/action"
	"wsf/controller/action/helper"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
	"wsf/utils"
)

const (
	// TYPEDefault represents default dispatcher
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

func init() {
	Register(TYPEDefault, NewDefaultDispatcher)
}

// Interface is a dispatcher interface
type Interface interface {
	SetOptions(options *Config) error
	Options() *Config
	SetLogger(l *log.Log) error
	Logger() *log.Log
	Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	IsDispatchable(rqs request.Interface) bool
	DefaultModule() string
	DefaultController() string
	DefaultAction() string
	RequestController(req request.Interface) (string, error)
	ActionMethod(req request.Interface) string
	SetModulesHandler(mds modules.Handler) error
	ModulesHandler() modules.Handler
	SetParams(params map[string]interface{}) error
	SetParam(name string, value interface{}) error
	Param(name string) interface{}
	ParamString(name string) string
	ParamBool(name string) bool
	Params() map[string]interface{}
	ClearParam(name string) bool
	ClearParams() bool
}

// Default dispatcher
type Default struct {
	options      *Config
	logger       *log.Log
	modules      modules.Handler
	invokeParams map[string]interface{}
}

// SetOptions sets dispatcher configuration
func (d *Default) SetOptions(options *Config) error {
	d.options = options
	return nil
}

// Options returns dispatcher configuration
func (d *Default) Options() *Config {
	return d.options
}

// Dispatch dispatches the request into the apropriet handler
func (d *Default) Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	md := d.modules.Module(rqs.ModuleName())
	if md == nil {
		if md = d.modules.Module(d.DefaultModule()); md == nil {
			return true, errors.Errorf("Invalid module specified '%s'", rqs.ModuleName())
		}
	}

	ctrl, err := md.Controller(rqs.ControllerName())
	if err != nil {
		if ctrl, err = md.Controller(d.DefaultController()); err != nil {
			return true, err
		}
	}

	act := d.ActionMethod(rqs)
	mtd, ok := reflect.TypeOf(ctrl).MethodByName(act)
	if !ok {
		return true, errors.Errorf("Action '%s' does not exists", act)
	}

	if !d.ParamBool("noViewRenderer") && !ctrl.HelperBroker().HasHelper("viewRenderer") {
		vr, err := helper.NewViewRenderer()
		if err != nil {
			return true, err
		}

		err = ctrl.HelperBroker().SetHelper(-80, vr, nil)
		if err != nil {
			return true, err
		}
	}

	// Initiate action controller
	rqs.SetDispatched(true)
	if err = ctrl.Init(ctx); err != nil {
		return true, err
	}

	err = ctrl.Dispatch(ctx, ctrl, mtd)
	if err != nil {
		return true, err
	}

	return true, nil
}

// IsDispatchable return true if request may be dispatched
func (d *Default) IsDispatchable(rqs request.Interface) bool {
	ctrl, err := d.RequestController(rqs)
	if err != nil {
		return false
	}

	if ctrl == "" {
		return false
	}

	return true
}

// DefaultModule returns default dispatcher module
func (d *Default) DefaultModule() string {
	return d.options.defaultModule
}

// DefaultController returns default dispatcher controller
func (d *Default) DefaultController() string {
	return d.options.defaultController
}

// DefaultAction returns default dispatcher action
func (d *Default) DefaultAction() string {
	return d.options.defaultAction
}

// Handler returns request specific handler
func (d *Default) Handler(rqs request.Interface, rsp response.Interface) string {
	return rqs.ControllerName()
}

// RequestController returns controller name based on request
func (d *Default) RequestController(req request.Interface) (string, error) {
	controllerName := req.ControllerName()
	if controllerName == "" {
		if !d.options.useDefaultControllerAlways {
			return "", nil
		}

		controllerName = d.DefaultController()
		req.SetControllerName(controllerName)
	}

	controllerName = d.formatControllerName(controllerName)
	module := req.ModuleName()
	if !d.IsValidModule(module) {
		return "", errors.New("No default module defined for this application")
	}

	return controllerName, nil
}

// IsValidModule returns true if provided module is registered
func (d *Default) IsValidModule(md string) bool {
	return true
}

// ActionMethod returns action name from request
func (d *Default) ActionMethod(req request.Interface) string {
	action := req.ActionName()
	if action == "" {
		action = d.DefaultAction()
		req.SetActionName(action)
	}

	return d.formatActionName(action)
}

// SetModulesHandler sets module handler for this dispatcher
func (d *Default) SetModulesHandler(mds modules.Handler) error {
	d.modules = mds
	return nil
}

// ModulesHandler returns dispatcher modules handler
func (d *Default) ModulesHandler() modules.Handler {
	return d.modules
}

// SetParams sets parameters to pass to handlers
func (d *Default) SetParams(params map[string]interface{}) error {
	d.invokeParams = utils.MapSMerge(d.invokeParams, params)
	return nil
}

// SetParam add or modify a parameter to use when instantiating a handler
func (d *Default) SetParam(name string, value interface{}) error {
	d.invokeParams[name] = value
	return nil
}

// Param retrieve a single parameter from the parameter stack
func (d *Default) Param(name string) interface{} {
	if v, ok := d.invokeParams[name]; ok {
		return v
	}

	return nil
}

// ParamString retrieve a single parameter from the parameter stack
func (d *Default) ParamString(name string) string {
	if v, ok := d.invokeParams[name]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// ParamBool retrieve a single parameter from the parameter stack
func (d *Default) ParamBool(name string) bool {
	if v, ok := d.invokeParams[name]; ok {
		if v, ok := v.(bool); ok {
			return v
		}

		return false
	}

	return false
}

// Params retrieve handler parameters
func (d *Default) Params() map[string]interface{} {
	return d.invokeParams
}

// ClearParam clears the specified parameter
func (d *Default) ClearParam(name string) bool {
	if _, ok := d.invokeParams[name]; ok {
		delete(d.invokeParams, name)
		return true
	}

	return false
}

// ClearParams clears the parameter stack
func (d *Default) ClearParams() bool {
	d.invokeParams = make(map[string]interface{})
	return true
}

// SetLogger attaches log writer
func (d *Default) SetLogger(l *log.Log) error {
	d.logger = l
	return nil
}

// Logger returns attached log writer
func (d *Default) Logger() *log.Log {
	return d.logger
}

func (d *Default) formatControllerName(name string) string {
	return name
}

func (d *Default) formatActionName(name string) string {
	parts := strings.Split(name, "-")
	for k, v := range parts {
		v = strings.ToLower(v)
		parts[k] = strings.Title(v)
	}

	return strings.Join(parts, "")
}

func (d *Default) findControllers(controllerType reflect.Type) (indexes [][]int) {
	controllerPtrType := reflect.TypeOf(&action.Controller{})

	// It might be a multi-level embedding. To find the controllers, we follow
	// every anonymous field, using breadth-first search.
	type nodeType struct {
		val   reflect.Value
		index []int
	}
	controllerPtr := reflect.New(controllerType)
	queue := []nodeType{{controllerPtr, []int{}}}
	for len(queue) > 0 {
		// Get the next value and de-reference it if necessary.
		var (
			node     = queue[0]
			elem     = node.val
			elemType = elem.Type()
		)

		if elemType.Kind() == reflect.Ptr {
			elem = elem.Elem()
			elemType = elem.Type()
		}
		queue = queue[1:]
		// #944 if the type's Kind is not `Struct` move on,
		// otherwise `elem.NumField()` will panic
		if elemType.Kind() != reflect.Struct {
			continue
		}

		// Look at all the struct fields.
		for i := 0; i < elem.NumField(); i++ {
			// If this is not an anonymous field, skip it.
			structField := elemType.Field(i)
			if !structField.Anonymous {
				continue
			}

			fieldValue := elem.Field(i)
			fieldType := structField.Type

			// If it's a Controller, record the field indexes to get here.
			if fieldType == controllerPtrType {
				indexes = append(indexes, append(node.index, i))
				continue
			}

			queue = append(queue, nodeType{fieldValue, append(append([]int{}, node.index...), i)})
		}
	}

	return
}

// NewDefaultDispatcher creates new default dispatcher
func NewDefaultDispatcher(options *Config) (di Interface, err error) {
	d := &Default{}
	d.SetOptions(options)

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return nil, errors.New("[Dispatcher] Log resource is required")
	}
	d.SetLogger(logResource.(*log.Log))

	return d, nil
}

// NewDispatcher creates a new dispatcher specified by type
func NewDispatcher(dispatcherType string, options *Config) (Interface, error) {
	if f, ok := buildHandlers[dispatcherType]; ok {
		dsp, err := f(options)
		if err != nil {
			return nil, err
		}

		if mds := registry.GetResource("modules"); mds != nil {
			dsp.SetModulesHandler(registry.GetResource("modules").(modules.Handler))
		}

		return dsp, nil
	}

	return nil, errors.Errorf("Unrecognized dispatcher type \"%v\"", dispatcherType)
}

// Register registers a handler for dispatcher creation
func Register(dispatcherType string, handler func(*Config) (Interface, error)) {
	buildHandlers[dispatcherType] = handler
}
