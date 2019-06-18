package bootstrap

import (
	"sync"
	"wsf/application/resource"
	"wsf/application/service"
	"wsf/config"
	"wsf/errors"
	"wsf/log"
)

const (
	// EventInfo thrown if there is something to say
	EventInfo = iota + 500

	// EventError thrown on any non job error provided
	EventError
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface Bootstrap interface
type Interface interface {
	Options() *Config
	SetOptions(*Config) error
	Init() error
	Run() error
	Stop()
	//RegisterResource(resource string, options config.Config) error
	//UnregisterResource(resource string) (i Interface, err error)
	HasResource(resource string) bool
	Resource(resource string) (r interface{}, status int)
	AddListener(l func(event int, ctx interface{}))
	Listen(l func(event int, ctx interface{}))
}

// Bootstrap is an abstract bootstrap struct
type Bootstrap struct {
	options   *Config
	Resources resource.Registry
	Services  service.Server
	logger    *log.Log
	lsns      []func(event int, ctx interface{})
	lsn       func(event int, ctx interface{})
	mu        sync.Mutex
}

// Init initializes the application
func (b *Bootstrap) Init() error {
	cfg := b.options.Get("resources")
	if cfg == nil {
		return errors.New("[Application] Resources configuration undefined")
	}

	if err := b.Resources.Init(cfg); err != nil {
		return err
	}

	cfg = b.options.Get("services")
	if cfg == nil {
		return errors.New("[Application] Services configuration undefined")
	}

	if err := b.Services.Init(cfg); err != nil {
		return err
	}

	return nil
}

// Run Serves the application
func (b *Bootstrap) Run() error {
	return b.Services.Serve()
}

// Stop stops the application
func (b *Bootstrap) Stop() {
	b.Services.Stop()
}

// Options returns configuration of the bootstrap struct
func (b *Bootstrap) Options() *Config {
	return b.options
}

// SetOptions Sets configuration for bootsrap struct
func (b *Bootstrap) SetOptions(options *Config) error {
	if options == nil {
		return nil
	}

	b.options = options
	return nil
}

// HasResource returns true if resource is registered
func (b *Bootstrap) HasResource(name string) bool {
	return b.Resources.Has(name)
}

// Resource returns resource by name
func (b *Bootstrap) Resource(name string) (interface{}, int) {
	return b.Resources.Get(name)
}

// HasService returns true if service is registered
func (b *Bootstrap) HasService(name string) bool {
	return b.Services.Has(name)
}

// Service returns service by name
func (b *Bootstrap) Service(name string) (interface{}, int) {
	return b.Services.Get(name)
}

// AddListener attaches event watcher
func (b *Bootstrap) AddListener(l func(event int, ctx interface{})) {
	b.lsns = append(b.lsns, l)
}

// Listen attaches handler event watcher
func (b *Bootstrap) Listen(l func(event int, ctx interface{})) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lsn = l
}

// throw invokes event handler if any
func (b *Bootstrap) throw(event int, ctx interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.lsn != nil {
		b.lsn(event, ctx)
	}
}

// NewBootstrap Creates boostrap struct
func NewBootstrap(options config.Config) (b Interface, err error) {
	if options == nil {
		return nil, errors.Errorf("Application configuration can not be empty")
	}

	bcfg := &Config{AppConfig: options}
	cfg := options.Get("application.bootstrap")
	if cfg == nil {
		bcfg.Defaults()
	} else {
		bcfg.Populate(cfg)
	}

	if f, ok := buildHandlers[bcfg.Type]; ok {
		return f(bcfg)
	}

	return nil, errors.Errorf("Unrecognized bootstrap type \"%v\"", bcfg.Type)
}

// Register registers a handler for bootstrap creation
func Register(bootstrapType string, handler func(*Config) (Interface, error)) {
	buildHandlers[bootstrapType] = handler
}
