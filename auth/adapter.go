package auth

import (
	"wsf/config"
	"wsf/context"
	"wsf/errors"
)

var (
	buildAdapterHandlers = map[string]func(*AdapterConfig) (Adapter, error){}
)

// Adapter represents auth adapter interface
type Adapter interface {
	Setup() error
	Authenticate(ctx context.Context) Result
}

// NewAdapter creates a new auth adapter from given type and options
func NewAdapter(adapterType string, options config.Config) (Adapter, error) {
	cfg := &AdapterConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildAdapterHandlers[adapterType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized auth adapter type \"%v\"", adapterType)
}

// NewAdapterFromConfig creates a new auth adapter from given type and AdapterConfig
func NewAdapterFromConfig(adapterType string, options *AdapterConfig) (Adapter, error) {
	if f, ok := buildAdapterHandlers[adapterType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized auth adapter type \"%v\"", adapterType)
}

// RegisterAdapter registers a handler for auth adapter creation
func RegisterAdapter(adapterType string, handler func(*AdapterConfig) (Adapter, error)) {
	buildAdapterHandlers[adapterType] = handler
}
