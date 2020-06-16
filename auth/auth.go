package auth

import (
	"sync"
	"wsf/config"
	"wsf/context"
	"wsf/errors"
)

// Public constants
const (
	TYPEDefault = "default"

	ROLEGuest = "guest"
	ROLEUser  = "user"
	ROLEAdmin = "admin"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}

	ins Interface
)

func init() {
	Register(TYPEDefault, NewDefaultAuth)
}

// Interface is an auth interface
type Interface interface {
	Priority() int
	Init(options *Config) (bool, error)
	GetOptions() *Config
	SetStorage(strg Storage) error
	GetStorage() Storage
	Authenticate(ctx context.Context, adp Adapter) Result
	HasIdentity(ctx context.Context) bool
	Identity(ctx context.Context) Identity
	ClearIdentity(ctx context.Context) bool
	ClearIdentityes() bool
	CreateIdentity(data map[string]interface{}) (Identity, error)
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

// Init the object
func (a *DefaultAuth) Init(options *Config) (ok bool, err error) {
	a.Options = options
	a.Adapter, err = NewAdapterFromConfig(a.Options.Adapter.Type, a.Options.Adapter)
	if err != nil {
		return ok, err
	}

	a.Storage, err = NewStorageFromConfig(a.Options.Storage.Type, a.Options.Storage)
	if err != nil {
		return ok, err
	}

	return true, nil
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
	var result Result
	if adp != nil {
		result = adp.Authenticate(ctx)
	} else {
		result = a.Adapter.Authenticate(ctx)
	}

	if a.HasIdentity(ctx) {
		a.ClearIdentity(ctx)
	}

	if result.Valid() {
		a.GetStorage().Write(ctx, result.GetIdentity().Map())
	}

	return result
}

// HasIdentity returns true if and only if an identity is available from storage
func (a *DefaultAuth) HasIdentity(ctx context.Context) bool {
	return !a.GetStorage().IsEmpty(ctx)
}

// Identity returns the identity from storage or null if no identity is available
func (a *DefaultAuth) Identity(ctx context.Context) Identity {
	if a.GetStorage().IsEmpty(ctx) {
		return a.Guest()
	}

	data, err := a.GetStorage().Read(ctx)
	if err != nil {
		return a.Guest()
	}

	idnt, err := a.CreateIdentity(data)
	if err != nil {
		return a.Guest()
	}

	return idnt
}

// ClearIdentity clears the identity from persistent storage
func (a *DefaultAuth) ClearIdentity(ctx context.Context) bool {
	return a.GetStorage().Clear(ctx)
}

// ClearIdentityes clears the storage
func (a *DefaultAuth) ClearIdentityes() bool {
	return a.GetStorage().ClearAll()
}

// CreateIdentity creates auth specific identity
func (a *DefaultAuth) CreateIdentity(data map[string]interface{}) (Identity, error) {
	return NewIdentityFromConfig(Options().Identity, data)
}

// Guest returns new guest identity
func (a *DefaultAuth) Guest() Identity {
	idnt, err := NewIdentityFromConfig(a.Options.Identity, map[string]interface{}{
		"id":         0,
		"role":       ROLEGuest,
		"roleID":     0,
		"instanceID": 0,
		"name":       "Guest",
	})
	if err != nil {
		idnt, _ := NewIdentityDefault(a.Options.Identity, map[string]interface{}{
			"id":         0,
			"role":       ROLEGuest,
			"roleID":     0,
			"instanceID": 0,
			"name":       "Guest",
		})
		return idnt
	}

	return idnt
}

// NewDefaultAuth creates a default auth object
func NewDefaultAuth(options *Config) (Interface, error) {
	a := &DefaultAuth{}
	a.Options = options

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
