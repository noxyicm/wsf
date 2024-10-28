package resource

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/controller"
)

// TYPEController id of resource
const TYPEController = "controller"

func init() {
	Register(TYPEController, NewControllerResource)
}

// NewControllerResource creates a new resource of type Controller
func NewControllerResource(cfg config.Config) (Interface, error) {
	controllerType := cfg.GetString("type")
	ctrl, err := controller.NewController(controllerType, cfg)
	if err != nil {
		return nil, err
	}

	controller.SetInstance(ctrl)
	return ctrl, nil
}
