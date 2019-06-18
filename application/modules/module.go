package modules

import (
	"path/filepath"
	"reflect"
	"wsf/controller/action"
	"wsf/errors"
	"wsf/filter"
	"wsf/filter/word"
	"wsf/registry"
	"wsf/view"
)

// Module represents a module
type Module struct {
	Name         string
	Order        int
	ViewPathSpec string
	controllers  map[string]reflect.Type
	handlers     map[string]action.Interface
}

// RegisterController registers action controller
func (m *Module) RegisterController(controller string, controllerType reflect.Type) error {
	m.controllers[controller] = controllerType

	m.RegisterScriptPath(controller)
	return nil
}

// ControllerType returns controller type by its name
func (m *Module) ControllerType(name string) (reflect.Type, error) {
	if v, ok := m.controllers[name]; ok {
		return v, nil
	}

	return nil, errors.Errorf("Invalid controller specified (%s)", name)
}

// RegisterScriptPath registers a paths assosiated with controller
func (m *Module) RegisterScriptPath(controllerName string) error {
	inf, err := filter.NewInflector()
	if err != nil {
		return nil
	}

	inflector := inf.(*filter.Inflector)
	uts, err := word.NewUnderscoreToSeparator("/")
	if err != nil {
		return nil
	}

	rrc, err := filter.NewRegexpReplace(`\.`, "-")
	if err != nil {
		return nil
	}

	inflector.AddRules(map[string]interface{}{
		":module":     []interface{}{"Word_CamelCaseToDash", "StringToLower"},
		":controller": []interface{}{"Word_CamelCaseToDash", uts, "StringToLower", rrc},
	})

	inflector.SetTarget(m.ViewPathSpec)

	controllerPath, err := inflector.Filter(map[string]string{
		"module":     m.Name,
		"controller": controllerName,
	})
	if err != nil {
		return err
	}

	if v := registry.Get("view"); v != nil {
		v.(view.Interface).AddTeplatePath(filepath.FromSlash(controllerPath.(string) + "/"))
	}

	return nil
}

// GetResource returns a registered resource from registry
func (m *Module) GetResource(name string) interface{} {
	return registry.Get(name)
}

// NewModule creates new module struct
func NewModule(name string, order int) (*Module, error) {
	return &Module{
		Name:         name,
		Order:        order,
		ViewPathSpec: "views/:module/:controller",
		controllers:  make(map[string]reflect.Type),
		handlers:     make(map[string]action.Interface),
	}, nil
}
