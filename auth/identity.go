package auth

import (
	"context"
	"wsf/config"
	"wsf/errors"
)

// Public constants
const (
	TYPEIdentityDefault = "default"
)

type contextKey int

var (
	buildIdentityHandlers = map[string]func(*IdentityConfig, map[string]interface{}) (Identity, error){}
	identityContextKey    contextKey
)

func init() {
	RegisterIdentity(TYPEIdentityDefault, NewIdentityDefault)
}

// Identity represents auth identity interface
type Identity interface {
	Setup() error
	Set(map[string]interface{})
	Get(key string) interface{}
	GetInt(key string) int
}

// NewIdentity creates a new auth identity from given type and options
func NewIdentity(identityType string, options config.Config, data map[string]interface{}) (Identity, error) {
	cfg := &IdentityConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildIdentityHandlers[identityType]; ok {
		return f(cfg, data)
	}

	return nil, errors.Errorf("Unrecognized auth identity type \"%v\"", identityType)
}

// NewIdentityFromConfig creates a new auth identity from given IdentityConfig
func NewIdentityFromConfig(options *IdentityConfig, data map[string]interface{}) (Identity, error) {
	identityType := options.Type
	if f, ok := buildIdentityHandlers[identityType]; ok {
		return f(options, data)
	}

	return nil, errors.Errorf("Unrecognized auth identity type \"%v\"", identityType)
}

// RegisterIdentity registers a handler for auth identity creation
func RegisterIdentity(identityType string, handler func(*IdentityConfig, map[string]interface{}) (Identity, error)) {
	buildIdentityHandlers[identityType] = handler
}

// DefaultIdentity is a default auth identity
type DefaultIdentity struct {
	Options *IdentityConfig
	Data    map[string]interface{}
}

// Setup the object
func (i *DefaultIdentity) Setup() error {
	return nil
}

// Set identity data
func (i *DefaultIdentity) Set(m map[string]interface{}) {
	i.Data = m
}

// Get returns an identity value by its key
func (i *DefaultIdentity) Get(key string) interface{} {
	return i.Data[key]
}

// GetInt returns an identity value by its key as int
func (i *DefaultIdentity) GetInt(key string) int {
	return i.Data[key].(int)
}

// NewIdentityDefault creates a new default auth identity
func NewIdentityDefault(options *IdentityConfig, data map[string]interface{}) (Identity, error) {
	i := &DefaultIdentity{}
	i.Options = options
	i.Data = data
	i.Setup()

	return i, nil
}

// IdentityToContext returns a new context with stored identity
func IdentityToContext(ctx context.Context, idnt Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, idnt)
}

// IdentityFromContext returns an identity stored in context
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	v, ok := ctx.Value(identityContextKey).(Identity)
	return v, ok
}
