package resource

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/translate"
)

// TYPETranslate id of resource
const TYPETranslate = "translate"

func init() {
	Register(TYPETranslate, NewTranslateResource)
}

// NewTranslateResource creates a new resource of type Translate
func NewTranslateResource(cfg config.Config) (Interface, error) {
	typ := cfg.GetString("type")
	a, err := translate.NewTranslate(typ, cfg)
	if err != nil {
		return nil, err
	}

	translate.SetInstance(a)
	return a, nil
}
