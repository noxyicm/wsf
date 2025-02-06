package controller

import (
	"reflect"
	"strings"

	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view"
)

const (
	// TYPEDispatcherDefault represents default dispatcher
	TYPEDispatcherDefault = "default"
)

var (
	buildDispatcherHandlers = map[string]func(*DispatcherConfig) (DispatcherInterface, error){}
)

func init() {
	RegisterDispatcher(TYPEDispatcherDefault, NewDefaultDispatcher)
}

// DispatcherInterface is a dispatcher interface
type DispatcherInterface interface {
	SetOptions(options *DispatcherConfig) error
	Options() *DispatcherConfig
	SetLogger(l *log.Log) error
	Logger() *log.Log
	AddActionController(moduleName string, controllerName string, cnstr func() (ActionControllerInterface, error)) error
	Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error)
	IsDispatchable(rqs request.Interface) bool
	DefaultModule() string
	DefaultController() string
	DefaultAction() string
	RequestController(req request.Interface) (string, error)
	ActionMethod(req request.Interface) string
	//SetModulesHandler(mds modules.Handler) error
	//ModulesHandler() modules.Handler
	SetParams(params map[string]interface{}) error
	SetParam(name string, value interface{}) error
	Param(name string) interface{}
	ParamString(name string) string
	ParamBool(name string) bool
	Params() map[string]interface{}
	ClearParam(name string) bool
	ClearParams() bool
}

// DefaultDispatcher dispatcher
type DefaultDispatcher struct {
	options *DispatcherConfig
	logger  *log.Log
	//modules      modules.Handler
	actions      map[string]map[string]func() (ActionControllerInterface, error)
	invokeParams map[string]interface{}
}

// SetOptions sets dispatcher configuration
func (d *DefaultDispatcher) SetOptions(options *DispatcherConfig) error {
	d.options = options
	return nil
}

// Options returns dispatcher configuration
func (d *DefaultDispatcher) Options() *DispatcherConfig {
	return d.options
}

// AddActionController add controller for action handling
func (d *DefaultDispatcher) AddActionController(moduleName string, controllerName string, cnstr func() (ActionControllerInterface, error)) error {
	if m, ok := d.actions[moduleName]; ok {
		if _, ok := m[controllerName]; ok {
			return errors.Errorf("Controller '%s' in module '%s' already registered", controllerName, moduleName)
		}

		m[controllerName] = cnstr
		return nil
	}

	d.actions[moduleName] = map[string]func() (ActionControllerInterface, error){
		controllerName: cnstr,
	}

	return nil
}

// Dispatch dispatches the request into the apropriet handler
func (d *DefaultDispatcher) Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	md, ok := d.actions[rqs.ModuleName()]
	if !ok || md == nil {
		if md, ok = d.actions[d.DefaultModule()]; !ok || md == nil {
			return true, errors.Errorf("Invalid module specified '%s'", rqs.ModuleName())
		}
	}

	cnstr, ok := md[rqs.ControllerName()]
	if !ok {
		if cnstr, ok = md[d.DefaultController()]; !ok {
			return true, errors.Errorf("Invalid controller specified '%s' for module '%s'", rqs.ControllerName(), rqs.ModuleName())
		}
	}

	ctrl, err := cnstr()
	if err != nil {
		return true, errors.Wrapf(err, "Unable to instantiate controller '%s' for module '%s'", rqs.ControllerName(), rqs.ModuleName())
	}

	act := d.ActionMethod(rqs)
	mtd, ok := reflect.TypeOf(ctrl).MethodByName(act)
	if !ok {
		return true, errors.Errorf("Action '%s' does not exists", act)
	}

	vw := registry.GetResource("view")
	if vw != nil {
		if err := ctrl.SetView(vw.(view.Interface)); err != nil {
			d.Logger().Notice(err, nil)
		}
	}

	ctrl.SetHelperBroker(GetHelperBroker())
	if !d.ParamBool("noViewRenderer") && !ctrl.HelperBroker().HasHelper("ViewRenderer") {
		vr, err := NewViewRendererHelper()
		if err != nil {
			return true, err
		}

		err = ctrl.HelperBroker().SetHelper(-80, vr, false, nil)
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
func (d *DefaultDispatcher) IsDispatchable(rqs request.Interface) bool {
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
func (d *DefaultDispatcher) DefaultModule() string {
	return d.options.defaultModule
}

// DefaultController returns default dispatcher controller
func (d *DefaultDispatcher) DefaultController() string {
	return d.options.defaultController
}

// DefaultAction returns default dispatcher action
func (d *DefaultDispatcher) DefaultAction() string {
	return d.options.defaultAction
}

// Handler returns request specific handler
func (d *DefaultDispatcher) Handler(rqs request.Interface, rsp response.Interface) string {
	return rqs.ControllerName()
}

// RequestController returns controller name based on request
func (d *DefaultDispatcher) RequestController(req request.Interface) (string, error) {
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
func (d *DefaultDispatcher) IsValidModule(md string) bool {
	return true
}

// ActionMethod returns action name from request
func (d *DefaultDispatcher) ActionMethod(req request.Interface) string {
	action := req.ActionName()
	if action == "" {
		action = d.DefaultAction()
		req.SetActionName(action)
	}

	return d.formatActionName(action)
}

// SetModulesHandler sets module handler for this dispatcher
//func (d *DefaultDispatcher) SetModulesHandler(mds modules.Handler) error {
//	d.modules = mds
//	return nil
//}

// ModulesHandler returns dispatcher modules handler
//func (d *DefaultDispatcher) ModulesHandler() modules.Handler {
//	return d.modules
//}

// SetParams sets parameters to pass to handlers
func (d *DefaultDispatcher) SetParams(params map[string]interface{}) error {
	d.invokeParams = utils.MapSMerge(d.invokeParams, params)
	return nil
}

// SetParam add or modify a parameter to use when instantiating a handler
func (d *DefaultDispatcher) SetParam(name string, value interface{}) error {
	d.invokeParams[name] = value
	return nil
}

// Param retrieve a single parameter from the parameter stack
func (d *DefaultDispatcher) Param(name string) interface{} {
	if v, ok := d.invokeParams[name]; ok {
		return v
	}

	return nil
}

// ParamString retrieve a single parameter from the parameter stack
func (d *DefaultDispatcher) ParamString(name string) string {
	if v, ok := d.invokeParams[name]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// ParamBool retrieve a single parameter from the parameter stack
func (d *DefaultDispatcher) ParamBool(name string) bool {
	if v, ok := d.invokeParams[name]; ok {
		if v, ok := v.(bool); ok {
			return v
		}

		return false
	}

	return false
}

// Params retrieve handler parameters
func (d *DefaultDispatcher) Params() map[string]interface{} {
	return d.invokeParams
}

// ClearParam clears the specified parameter
func (d *DefaultDispatcher) ClearParam(name string) bool {
	if _, ok := d.invokeParams[name]; ok {
		delete(d.invokeParams, name)
		return true
	}

	return false
}

// ClearParams clears the parameter stack
func (d *DefaultDispatcher) ClearParams() bool {
	d.invokeParams = make(map[string]interface{})
	return true
}

// SetLogger attaches log writer
func (d *DefaultDispatcher) SetLogger(l *log.Log) error {
	d.logger = l
	return nil
}

// Logger returns attached log writer
func (d *DefaultDispatcher) Logger() *log.Log {
	return d.logger
}

func (d *DefaultDispatcher) formatControllerName(name string) string {
	return name
}

func (d *DefaultDispatcher) formatActionName(name string) string {
	parts := strings.Split(name, "-")
	for k, v := range parts {
		v = strings.ToLower(v)
		parts[k] = strings.Title(v)
	}

	return strings.Join(parts, "")
}

// func (d *DefaultDispatcher) findControllers(controllerType reflect.Type) (indexes [][]int) {
// 	controllerPtrType := reflect.TypeOf(&ActionControllerBase{})

// 	// It might be a multi-level embedding. To find the controllers, we follow
// 	// every anonymous field, using breadth-first search.
// 	type nodeType struct {
// 		val   reflect.Value
// 		index []int
// 	}
// 	controllerPtr := reflect.New(controllerType)
// 	queue := []nodeType{{controllerPtr, []int{}}}
// 	for len(queue) > 0 {
// 		// Get the next value and de-reference it if necessary.
// 		var (
// 			node     = queue[0]
// 			elem     = node.val
// 			elemType = elem.Type()
// 		)

// 		if elemType.Kind() == reflect.Ptr {
// 			elem = elem.Elem()
// 			elemType = elem.Type()
// 		}
// 		queue = queue[1:]
// 		// if the type's Kind is not `Struct` move on,
// 		// otherwise `elem.NumField()` will panic
// 		if elemType.Kind() != reflect.Struct {
// 			continue
// 		}

// 		// Look at all the struct fields.
// 		for i := 0; i < elem.NumField(); i++ {
// 			// If this is not an anonymous field, skip it.
// 			structField := elemType.Field(i)
// 			if !structField.Anonymous {
// 				continue
// 			}

// 			fieldValue := elem.Field(i)
// 			fieldType := structField.Type

// 			// If it's a Controller, record the field indexes to get here.
// 			if fieldType == controllerPtrType {
// 				indexes = append(indexes, append(node.index, i))
// 				continue
// 			}

// 			queue = append(queue, nodeType{fieldValue, append(append([]int{}, node.index...), i)})
// 		}
// 	}

// 	return
// }

// NewDefaultDispatcher creates new default dispatcher
func NewDefaultDispatcher(options *DispatcherConfig) (di DispatcherInterface, err error) {
	d := &DefaultDispatcher{
		actions: make(map[string]map[string]func() (ActionControllerInterface, error)),
	}
	d.SetOptions(options)

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return nil, errors.New("[Dispatcher] Log resource is required")
	}
	d.SetLogger(logResource.(*log.Log))

	return d, nil
}

// NewDispatcher creates a new dispatcher specified by type
func NewDispatcher(dispatcherType string, options *DispatcherConfig) (DispatcherInterface, error) {
	if f, ok := buildDispatcherHandlers[dispatcherType]; ok {
		dsp, err := f(options)
		if err != nil {
			return nil, err
		}

		//if mds := registry.GetResource("modules"); mds != nil {
		//	dsp.SetModulesHandler(registry.GetResource("modules").(modules.Handler))
		//}

		return dsp, nil
	}

	return nil, errors.Errorf("Unrecognized dispatcher type \"%v\"", dispatcherType)
}

// RegisterDispatcher registers a handler for dispatcher creation
func RegisterDispatcher(dispatcherType string, handler func(*DispatcherConfig) (DispatcherInterface, error)) {
	buildDispatcherHandlers[dispatcherType] = handler
}
