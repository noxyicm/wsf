package resource

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/layout"
)

// TYPELayout id of resource
const TYPELayout = "layout"

func init() {
	Register(TYPELayout, NewLayoutResource)
}

// NewLayoutResource creates a new resource of type Layout
func NewLayoutResource(cfg config.Config) (Interface, error) {
	resourceType := cfg.GetString("type")
	rsr, err := layout.NewLayout(resourceType, cfg)
	if err != nil {
		return nil, err
	}

	return rsr, nil
}
