package helper

import (
	"wsf/context"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/session"
	"wsf/view"
)

var (
	buildHandlers = map[string]func() (Interface, error){}
)

// Interface represents action helper interface
type Interface interface {
	Name() string
	Init(options map[string]interface{}) error
	PreDispatch(ctx context.Context) error
	PostDispatch(ctx context.Context) error
}

// Abstract is a base for action helpers
type Abstract struct {
	name             string
	actionController ControllerInterface
}

// Name returns helper name
func (h *Abstract) Name() string {
	return h.name
}

// Init the helper
func (h *Abstract) Init(options map[string]interface{}) error {
	return nil
}

// PreDispatch do dispatch preparations
func (h *Abstract) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *Abstract) PostDispatch(ctx context.Context) error {
	return nil
}

// NewHelper creates a new action helper specified by type
func NewHelper(helperType string) (Interface, error) {
	if f, ok := buildHandlers[helperType]; ok {
		return f()
	}

	return nil, errors.Errorf("Unrecognized helper type \"%v\"", helperType)
}

// NewHelperAbstract as
func NewHelperAbstract(name string) *Abstract {
	return &Abstract{
		name: name,
	}
}

// Register registers a handler for action helper creation
func Register(helperType string, handler func() (Interface, error)) {
	buildHandlers[helperType] = handler
}

// ControllerInterface is
type ControllerInterface interface {
	Request() request.Interface
	Response() response.Interface
	SetParams(params map[string]interface{}) error
	SetParam(name string, value interface{}) error
	Param(name string) interface{}
	ParamString(name string) string
	ParamBool(name string) bool
	Params() map[string]interface{}
	SetView(v view.Interface) error
	SetViewSuffix(suffix string) error
	View() view.Interface
	HasHelper(name string) bool
	Helper(name string) Interface
	SetContext(ctx context.Context)
	Context() context.Context
	SetSession(s session.Interface)
	Session() session.Interface
}
