package modules

import (
	"path/filepath"

	"github.com/noxyicm/wsf/controller"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/filter"
	"github.com/noxyicm/wsf/filter/word"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/view"
)

// Interface defines a module
type Interface interface {
	Name() string
	Order() int
	RegisterController(controllerName string, cnstr func() (controller.ActionControllerInterface, error)) error
	Controller(name string) (func() (controller.ActionControllerInterface, error), error)
	RegisterScriptPath(controllerName string) error
	Resource(name string) interface{}
	InitControllers() error
	InitAccess() error
	InitPlugins() error
	InitHelpers() error
	InitRoutes() error
	InitView() error
}

// Module represents a module
type Module struct {
	name         string
	order        int
	ViewPathSpec string
	controllers  map[string]func() (controller.ActionControllerInterface, error)
}

// Name returns module name
func (m *Module) Name() string {
	return m.name
}

// Order returns module order
func (m *Module) Order() int {
	return m.order
}

// RegisterController registers action controller
func (m *Module) RegisterController(controllerName string, cnstr func() (controller.ActionControllerInterface, error)) error {
	m.controllers[controllerName] = cnstr
	return controller.Dispatcher().AddActionController(m.name, controllerName, cnstr)
}

// Controller returns controller type by its name
func (m *Module) Controller(name string) (func() (controller.ActionControllerInterface, error), error) {
	if v, ok := m.controllers[name]; ok {
		return v, nil
	}

	return nil, errors.Errorf("Invalid controller specified (%s) for module '%s'", name, m.Name())
}

// RegisterScriptPath registers a paths assosiated with controller
func (m *Module) RegisterScriptPath(controllerName string) error {
	viewResource := registry.GetResource("view")
	if viewResource == nil {
		return errors.New("'view' resource has not been initialized")
	}

	v, ok := viewResource.(view.Interface)
	if !ok {
		return errors.New("View resource does not implements \"github.com/noxyicm/wsf/view\".Interface")
	}

	inflector, err := filter.NewInflector()
	if err != nil {
		return errors.Wrap(err, "Unable to add controller path for view templates")
	}

	uts, err := word.NewUnderscoreToSeparator("/")
	if err != nil {
		return errors.Wrap(err, "Unable to add controller path for view templates")
	}

	rrc, err := filter.NewRegexpReplace(`\.`, "-")
	if err != nil {
		return errors.Wrap(err, "Unable to add controller path for view templates")
	}

	inflector.AddRules(map[string]interface{}{
		":module":     []interface{}{"Word_CamelCaseToDash", "StringToLower"},
		":controller": []interface{}{"Word_CamelCaseToDash", uts, "StringToLower", rrc},
	})

	inflector.SetTarget(v.GetBasePath())

	controllerPath, err := inflector.Filter(map[string]string{
		"module":     m.name,
		"controller": controllerName,
	})
	if err != nil {
		return errors.Wrap(err, "Unable to add controller path for view templates")
	}

	v.AddTemplatePath(filepath.FromSlash(controllerPath.(string)))
	return nil
}

// Resource returns a registered resource from registry
func (m *Module) Resource(name string) interface{} {
	return registry.GetResource(name)
}

// InitControllers runs the controller initialization
func (m *Module) InitControllers() error {
	return nil
}

// InitAccess runs the access initialization
func (m *Module) InitAccess() error {
	return nil
}

// InitPlugins runs the controller plugins initialization
func (m *Module) InitPlugins() error {
	return nil
}

// InitHelpers runs the controller helpers initialization
func (m *Module) InitHelpers() error {
	return nil
}

// InitRoutes runs the controller initialization
func (m *Module) InitRoutes() error {
	return nil
}

// InitView runs the controller initialization
func (m *Module) InitView() error {
	return nil
}

// NewModule creates new module struct
func NewModule(order int, name string) *Module {
	return &Module{
		name:        name,
		order:       order,
		controllers: make(map[string]func() (controller.ActionControllerInterface, error)),
	}
}
