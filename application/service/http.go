package service

import (
	"wsf/config"
	"wsf/service"
	"wsf/service/http"
)

// TYPEHttp id of resource
const TYPEHttp = "http"

func init() {
	Register(TYPEHttp, NewHTTPService)
}

// NewHTTPService creates a new service of type Http
func NewHTTPService(cfg config.Config) (service.Interface, error) {
	svc, err := http.NewService(cfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
