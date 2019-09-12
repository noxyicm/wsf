package auth

import (
	"context"
	"sync"
	"wsf/config"
	"wsf/errors"
)

// Public constants
const (
	TYPEAuthDefault = "default"

	IdentityKey = "identity"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}

	ins Interface
)

func init() {
	Register(TYPEAuthDefault, NewDefaultAuth)
}

// Interface is an auth interface
type Interface interface {
	Setup() error
	Priority() int
	GetOptions() *Config
	SetStorage(strg Storage) error
	GetStorage() Storage
	Authenticate(ctx context.Context, adp Adapter) Result
	HasIdentity(idnt string) bool
	Identity(idnt string) (Identity, error)
	ClearIdentity(idnt string) bool
	ClearIdentityes() bool
}

// NewAuth creates a new auth from given type and options
func NewAuth(authType string, options config.Config) (i Interface, err error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[authType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized auth type \"%v\"", authType)
}

// Register registers a handler for auth creation
func Register(authType string, handler func(*Config) (Interface, error)) {
	buildHandlers[authType] = handler
}

// DefaultAuth is a default auth object
type DefaultAuth struct {
	Options *Config
	Adapter Adapter
	Storage Storage
	mu      sync.Mutex
}

// Priority returns a priority of resource
func (a *DefaultAuth) Priority() int {
	return a.Options.Priority
}

// Setup the object
func (a *DefaultAuth) Setup() (err error) {
	a.Adapter, err = NewAdapterFromConfig(a.Options.Adapter.Type, a.Options.Adapter)
	if err != nil {
		return err
	}

	a.Storage, err = NewStorageFromConfig(a.Options.Storage.Type, a.Options.Storage)
	if err != nil {
		return err
	}

	return nil
}

// GetOptions returns an auth options
func (a *DefaultAuth) GetOptions() *Config {
	return a.Options
}

// SetStorage sets a storage to auth object
func (a *DefaultAuth) SetStorage(strg Storage) error {
	a.Storage = strg
	return nil
}

// GetStorage returns an auth storage
func (a *DefaultAuth) GetStorage() Storage {
	return a.Storage
}

// Authenticate performs an authentication attempt
func (a *DefaultAuth) Authenticate(ctx context.Context, adp Adapter) Result {
	result := a.Adapter.Authenticate(ctx)
	idnt := ctx.Value(IdentityKey)

	if a.HasIdentity(idnt.(string)) {
		a.ClearIdentity(idnt.(string))
	}

	if result.Valid() {
		a.GetStorage().Write(idnt.(string), result.GetIdentity())
	}

	return result
}

// HasIdentity returns true if and only if an identity is available from storage
func (a *DefaultAuth) HasIdentity(idnt string) bool {
	return !a.GetStorage().IsEmpty(idnt)
}

// Identity returns the identity from storage or null if no identity is available
func (a *DefaultAuth) Identity(idnt string) (Identity, error) {
	storage := a.GetStorage()
	if storage.IsEmpty(idnt) {
		return nil, errors.Errorf("Storage does not contain identity '%s'", idnt)
	}

	return a.GetStorage().Read(idnt)
}

// ClearIdentity clears the identity from persistent storage
func (a *DefaultAuth) ClearIdentity(idnt string) bool {
	return a.GetStorage().Clear(idnt)
}

// ClearIdentityes clears the storage
func (a *DefaultAuth) ClearIdentityes() bool {
	return a.GetStorage().ClearAll()
}

// NewDefaultAuth creates a default auth object
func NewDefaultAuth(options *Config) (Interface, error) {
	a := &DefaultAuth{}
	a.Options = options
	a.Setup()

	return a, nil
}

// SetInstance sets a main auth instance
func SetInstance(a Interface) {
	ins = a
}

// Instance returns an auth instance
func Instance() Interface {
	return ins
}

// Options return db options
func Options() *Config {
	return ins.GetOptions()
}
