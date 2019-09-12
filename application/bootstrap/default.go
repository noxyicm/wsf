package bootstrap

import (
	"sync"
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

	return b, nil
}
