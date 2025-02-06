package view

import (
	"github.com/noxyicm/wsf/errors"
)

var (
	buildViewHelperHandlers = map[string]func() (HelperInterface, error){}
)

// HelperInterface represents action helper interface
type HelperInterface interface {
	Name() string
	Init(vi Interface, options map[string]interface{}) error
	Setup() error
	SetView(vi Interface) error
	Render() error
}

// AbstractHelper is a base for view helpers
type AbstractHelper struct {
	name string
}

// Name returns helper name
func (h *AbstractHelper) Name() string {
	return h.name
}

// Init the helper
func (h *AbstractHelper) Init(vi Interface, options map[string]interface{}) error {
	return nil
}

// Setup the helper
func (h *AbstractHelper) Setup() error {
	return nil
}

// SetView designates instance of view
func (h *AbstractHelper) SetView(vi Interface) error {
	return nil
}

// Render renders helper content
func (h *AbstractHelper) Render() error {
	return nil
}

// NewHelper creates a new view helper specified by type
func NewHelper(helperType string) (HelperInterface, error) {
	if f, ok := buildViewHelperHandlers[helperType]; ok {
		return f()
	}

	return nil, errors.Errorf("Unrecognized view helper type \"%v\"", helperType)
}

// NewHelperAbstract creates new instance of AbstractHelper
func NewHelperAbstract(name string) *AbstractHelper {
	return &AbstractHelper{
		name: name,
	}
}

// RegisterHelper registers a handler for view helper creation
func RegisterHelper(helperType string, handler func() (HelperInterface, error)) {
	buildViewHelperHandlers[helperType] = handler
}
