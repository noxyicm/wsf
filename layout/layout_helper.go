package layout

import (
	"wsf/controller/action/helper"
	"wsf/controller/context"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/session"
	"wsf/view"
)

// TYPELayoutActionHelper is a helper id
const TYPELayoutActionHelper = "layout"

func init() {
	helper.Register(TYPELayoutActionHelper, NewLayoutHelper)
}

// Helper is a layout action controller helper
type Helper struct {
	name             string
	actionController helper.ControllerInterface
	View             view.Interface
}

// Name returns helper name
func (h *Helper) Name() string {
	return h.name
}

// Init the helper
func (h *Helper) Init(options map[string]interface{}) error {
	return nil
}

// PreDispatch do dispatch preparations
func (h *Helper) PreDispatch() error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *Helper) PostDispatch() error {
	return nil
}

// SetController sets action controller
func (h *Helper) SetController(ctrl helper.ControllerInterface) error {
	h.actionController = ctrl
	return nil
}

// Controller returns action controller
func (h *Helper) Controller() helper.ControllerInterface {
	return h.actionController
}

// Request returns request object
func (h *Helper) Request() request.Interface {
	return h.Controller().Request()
}

// Response return response object
func (h *Helper) Response() response.Interface {
	return h.Controller().Response()
}

// Session return session object
func (h *Helper) Session() session.Interface {
	return h.Controller().Session()
}

// SetLayout sets a layout for render
func (h *Helper) SetLayout(name string) {
	h.Controller().Context().SetValue(context.Layout, name)
}

// Disable layout render
func (h *Helper) Disable() {
	h.Controller().Context().SetValue(context.LayoutEnabled, false)
}

// NewLayoutHelper creates a new layout action helper
func NewLayoutHelper() (helper.Interface, error) {
	return &Helper{
		name: TYPELayoutActionHelper,
	}, nil
}
