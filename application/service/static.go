package service

import (
	"wsf/config"
	"wsf/service"
	"wsf/service/static"
)

// TYPEStatic id of resource
const TYPEStatic = "static"

func init() {
	Register(TYPEStatic, NewStaticService)
}

// NewStaticService creates a new service of type Static
func NewStaticService(cfg config.Config) (service.Interface, error) {
	svc, err := static.NewService(cfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
