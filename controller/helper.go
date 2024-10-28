package controller

import (
	"wsf/context"
	"wsf/errors"
)

var (
	buildHelperHandlers = map[string]func() (HelperInterface, error){}
)

// HelperInterface represents action helper interface
type HelperInterface interface {
	Name() string
	Init(options map[string]interface{}) error
	PreDispatch(ctx context.Context) error
	PostDispatch(ctx context.Context) error
}

// AbstractHelper is a base for action helpers
type AbstractHelper struct {
	name string
}

// Name returns helper name
func (h *AbstractHelper) Name() string {
	return h.name
}

// Init the helper
func (h *AbstractHelper) Init(options map[string]interface{}) error {
	return nil
}

// PreDispatch do dispatch preparations
func (h *AbstractHelper) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *AbstractHelper) PostDispatch(ctx context.Context) error {
	return nil
}

// NewHelper creates a new action helper specified by type
func NewHelper(helperType string) (HelperInterface, error) {
	if f, ok := buildHelperHandlers[helperType]; ok {
		return f()
	}

	return nil, errors.Errorf("Unrecognized helper type \"%v\"", helperType)
}

// NewHelperAbstract creates new instance of AbstractHelper
func NewHelperAbstract(name string) *AbstractHelper {
	return &AbstractHelper{
		name: name,
	}
}

// RegisterHelper registers a handler for action helper creation
func RegisterHelper(helperType string, handler func() (HelperInterface, error)) {
	buildHelperHandlers[helperType] = handler
}
