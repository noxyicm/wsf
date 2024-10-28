package service

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/service"
	"github.com/noxyicm/wsf/service/tasker"
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
