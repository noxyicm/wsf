package layout

import (
	"wsf/controller"
	"wsf/view"
)

// TYPEHelperLayout is a helper id
const TYPEHelperLayout = "layout"

func init() {
	controller.RegisterHelper(TYPEHelperLayout, NewLayoutHelper)
}

// Helper is a layout action controller helper
type Helper struct {
	controller.AbstractHelper

	View view.Interface
}

// Name returns helper name
func (h *Helper) Name() string {
	return h.AbstractHelper.Name()
}

// NewLayoutHelper creates a new layout action helper
func NewLayoutHelper() (controller.HelperInterface, error) {
	return &Helper{
		AbstractHelper: *controller.NewHelperAbstract(TYPEHelperLayout),
	}, nil
}
