package service

import (
	"wsf/config"
	"wsf/service"
	"wsf/service/rest"
)

// TYPERest id of resource
const TYPERest = "rest"

func init() {
	Register(TYPERest, NewRESTService)
}

// NewRESTService creates a new service of type REST
func NewRESTService(cfg config.Config) (service.Interface, error) {
	svc, err := rest.NewService(cfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
