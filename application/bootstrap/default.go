package bootstrap

import (
	"sync"
	"wsf/application/resource"
	"wsf/application/service"
)

const (
	// TYPEDfault is a type of bootstrap
	TYPEDfault = "default"
)

func init() {
	Register(TYPEDfault, NewDefaultBootstrap)
}

// Default struct
type Default struct {
	Bootstrap

	mu sync.Mutex
}

// NewDefaultBootstrap creates boostrap struct
func NewDefaultBootstrap(options *Config) (Interface, error) {
	b := &Default{}
	b.SetOptions(options)

	b.Resources = resource.NewRegistry()
	b.Resources.Listen(b.throw)

	b.Services = service.NewServer()
	b.Services.Listen(b.throw)

	return b, nil
}
