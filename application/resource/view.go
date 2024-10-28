package resource

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/view"
)

// TYPEView id of resource
const TYPEView = "view"

func init() {
	Register(TYPEView, NewViewResource)
}

// NewViewResource creates a new resource of type View
func NewViewResource(cfg config.Config) (Interface, error) {
	viewType := cfg.GetString("type")
	v, err := view.NewView(viewType, cfg)
	if err != nil {
		return nil, err
	}

	return v, nil
}
