package service

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/service"
	"github.com/noxyicm/wsf/service/rpc"
)

// TYPERpc id of resource
const TYPERpc = "rpc"

func init() {
	Register(TYPERpc, NewRPCService)
}

// NewRPCService creates a new service of type Http
func NewRPCService(cfg config.Config) (service.Interface, error) {
	svc, err := rpc.NewService(cfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
