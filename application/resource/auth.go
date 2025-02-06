package resource

import (
	"github.com/noxyicm/wsf/auth"
	"github.com/noxyicm/wsf/config"
)

// TYPEAuth id of resource
const TYPEAuth = "auth"

func init() {
	Register(TYPEAuth, NewAuthResource)
}

// NewAuthResource creates a new resource of type Auth
func NewAuthResource(cfg config.Config) (Interface, error) {
	typ := cfg.GetString("type")
	a, err := auth.NewAuth(typ, cfg)
	if err != nil {
		return nil, err
	}

	auth.SetInstance(a)
	return a, nil
}
