package modules

import (
	"path/filepath"
	"wsf/controller/action"
	"wsf/errors"
	"wsf/filter"
	"wsf/filter/word"
	"wsf/registry"
	"wsf/view"
)

// Interface defines a module
type Interface interface {
	Name() string
	Order() int
	RegisterController(controllerName string, controller action.Interface) error
	Controller(name string) (action.Interface, error)
	RegisterScriptPath(controllerName string) error
	Resource(name string) interface{}
	InitControllers() error
	InitPlugins() error
	InitRoutes() error
	InitActionHelpers() error
}

// Module represents a module
type Module struct {
	name         string
	order        int
	ViewPathSpec string
	controllers  map[string]action.Interface
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
func (m *Module) RegisterController(controllerName string, controller action.Interface) error {
	m.controllers[controllerName] = controller

	m.RegisterScriptPath(controllerName)
	return nil
}

// Controller returns controller type by its name
func (m *Module) Controller(name string) (action.Interface, error) {
	if v, ok := m.controllers[name]; ok {
		return v, nil
	}

	return nil, errors.Errorf("Invalid controller specified (%s)", name)
}

// RegisterScriptPath registers a paths assosiated with controller
func (m *Module) RegisterScriptPath(controllerName string) error {
	viewResource := registry.GetResource("view")
	if viewResource == nil {
		return errors.New("View resource has not been initialized")
	}

	v, ok := viewResource.(view.Interface)
	if !ok {
		return errors.New("View resource does not implements \"wsf/view\".Interface")
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

// InitPlugins runs the controller initialization
func (m *Module) InitPlugins() error {
	return nil
}

// InitRoutes runs the controller initialization
func (m *Module) InitRoutes() error {
	return nil
}

// InitActionHelpers runs the controller initialization
func (m *Module) InitActionHelpers() error {
	return nil
}

// NewModule creates new module struct
func NewModule(order int, name string) *Module {
	return &Module{
		name:        name,
		order:       order,
		controllers: make(map[string]action.Interface),
	}
}
