package layout

import (
	"wsf/controller/action/helper"
	"wsf/view"
)

// TYPELayoutActionHelper is a helper id
const TYPELayoutActionHelper = "layout"

func init() {
	helper.Register(TYPELayoutActionHelper, NewLayoutHelper)
}

// Helper is a layout action controller helper
type Helper struct {
	helper.Abstract

	View view.Interface
}

// Name returns helper name
func (h *Helper) Name() string {
	return h.Abstract.Name()
}

// NewLayoutHelper creates a new layout action helper
func NewLayoutHelper() (helper.Interface, error) {
	return &Helper{
		Abstract: *helper.NewHelperAbstract(TYPELayoutActionHelper),
	}, nil
}
