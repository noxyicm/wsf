package resource

import (
	"github.com/noxyicm/wsf/application/modules"
	"github.com/noxyicm/wsf/config"
)

// TYPEModules id of resource
const TYPEModules = "modules"

func init() {
	Register(TYPEModules, NewModulesResource)
}

// NewModulesResource creates a new resource of type Module
func NewModulesResource(cfg config.Config) (Interface, error) {
	handlerType := cfg.GetString("type")
	mdl, err := modules.NewHandler(handlerType, cfg)
	if err != nil {
		return nil, err
	}

	return mdl, nil
}
