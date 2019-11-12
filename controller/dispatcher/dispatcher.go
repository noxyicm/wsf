package dispatcher

import (
	"reflect"
	"strings"
	"wsf/application/modules"
	"wsf/context"
	"wsf/controller/action"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
	"wsf/session"
	"wsf/utils"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface is a dispatcher interface
type Interface interface {
	Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	IsDispatchable(rqs request.Interface) bool
	DefaultModule() string
	DefaultController() string
	DefaultAction() string
	RequestController(req request.Interface) (string, error)
	PopulateController(ctx context.Context, ctrl interface{}, rqs request.Interface, rsp response.Interface, invokeArgs map[string]interface{}) error
	GetActionMethod(req request.Interface) (string, error)
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

type standart struct {
	options      *Config
	logger       *log.Log
	modules      modules.Handler
	invokeParams map[string]interface{}
}

func (d *standart) IsDispatchable(rqs request.Interface) bool {
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
func (d *standart) DefaultModule() string {
	return d.options.defaultModule
}

// DefaultController returns default dispatcher controller
func (d *standart) DefaultController() string {
	return d.options.defaultController
}

// DefaultAction returns default dispatcher action
func (d *standart) DefaultAction() string {
	return d.options.defaultAction
}

// Handler returns request specific handler
func (d *standart) Handler(rqs request.Interface, rsp response.Interface) string {
	return rqs.ControllerName()
}

// RequestController returns controller name based on request
func (d *standart) RequestController(req request.Interface) (string, error) {
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
func (d *standart) IsValidModule(md string) bool {
	return true
}

// GetActionMethod returns action name from request
func (d *standart) GetActionMethod(req request.Interface) (string, error) {
	action := req.ActionName()
	if action == "" {
		action = d.DefaultAction()
		req.SetActionName(action)
	}

	return d.formatActionName(action), nil
}

// SetModulesHandler sets module handler for this dispatcher
func (d *standart) SetModulesHandler(mds modules.Handler) error {
	d.modules = mds
	return nil
}

// ModulesHandler returns dispatcher modules handler
func (d *standart) ModulesHandler() modules.Handler {
	return d.modules
}

// SetParams sets parameters to pass to handlers
func (d *standart) SetParams(params map[string]interface{}) error {
	d.invokeParams = utils.MapSMerge(d.invokeParams, params)
	return nil
}

// SetParam add or modify a parameter to use when instantiating a handler
func (d *standart) SetParam(name string, value interface{}) error {
	d.invokeParams[name] = value
	return nil
}

// Param retrieve a single parameter from the parameter stack
func (d *standart) Param(name string) interface{} {
	if v, ok := d.invokeParams[name]; ok {
		return v
	}

	return nil
}

// Param retrieve a single parameter from the parameter stack
func (d *standart) ParamString(name string) string {
	if v, ok := d.invokeParams[name]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// ParamBool retrieve a single parameter from the parameter stack
func (d *standart) ParamBool(name string) bool {
	if v, ok := d.invokeParams[name]; ok {
		if v, ok := v.(bool); ok {
			return v
		}

		return false
	}

	return false
}

// Params retrieve handler parameters
func (d *standart) Params() map[string]interface{} {
	return d.invokeParams
}

// ClearParam clears the specified parameter
func (d *standart) ClearParam(name string) bool {
	if _, ok := d.invokeParams[name]; ok {
		delete(d.invokeParams, name)
		return true
	}

	return false
}

// ClearParams clears the parameter stack
func (d *standart) ClearParams() bool {
	d.invokeParams = make(map[string]interface{})
	return true
}

func (d *standart) formatControllerName(name string) string {
	return name
}

func (d *standart) formatActionName(name string) string {
	parts := strings.Split(name, "-")
	for k, v := range parts {
		v = strings.ToLower(v)
		parts[k] = strings.Title(v)
	}

	return strings.Join(parts, "")
}

// PopulateController populates action controller
func (d *standart) PopulateController(ctx context.Context, controller interface{}, rqs request.Interface, rsp response.Interface, invokeArgs map[string]interface{}) error {
	controllerIndexes := d.findControllers(reflect.TypeOf(controller).Elem())
	controllerValue := reflect.ValueOf(controller).Elem()
	actionController := &action.Controller{}

	actionController.SetRequest(rqs)
	actionController.SetResponse(rsp)
	actionController.SetContext(ctx)
	actionController.SetSession(ctx.Value(context.SessionKey).(session.Interface))
	actionController.SetParams(invokeArgs)

	value := reflect.ValueOf(actionController)
	for _, index := range controllerIndexes {
		controllerValue.FieldByIndex(index).Set(value)
	}

	return nil
}

func (d *standart) findControllers(controllerType reflect.Type) (indexes [][]int) {
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
