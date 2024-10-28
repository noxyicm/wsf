package service

import (
	"wsf/config"
	"wsf/service"
	"wsf/service/auth"
)

// TYPEAuth id of resource
const TYPEAuth = "auth"

func init() {
	Register(TYPEAuth, NewAuthService)
}

// NewAuthService creates a new service of type Auth
func NewAuthService(cfg config.Config) (service.Interface, error) {
	svc, err := auth.NewService(cfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
