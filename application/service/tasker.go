package service

import (
	"wsf/config"
	"wsf/service"
	"wsf/service/tasker"
)

// TYPETasker id of service
const TYPETasker = "tasker"

func init() {
	Register(TYPETasker, NewTaskerService)
}

// NewTaskerService creates a new service of type Static
func NewTaskerService(cfg config.Config) (service.Interface, error) {
	svc, err := tasker.NewService(cfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
