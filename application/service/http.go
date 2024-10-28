package service

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/service"
	"github.com/noxyicm/wsf/service/http"
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
