package bootstrap

import (
	"sync"
	"wsf/application/resource"
	"wsf/application/service"
	"wsf/config"
	"wsf/context"
	"wsf/errors"
	"wsf/log"
)

const (
	// EventDebug thrown if there is something insegnificant to say
	EventDebug = iota + 500

	// EventInfo thrown if there is something to say
	EventInfo

	// EventError thrown on any non job error provided
	EventError
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface Bootstrap interface
type Interface interface {
	SetOptions(*Config) error
	GetOptions() *Config
	Init() error
	Run(ctx context.Context) error
	Stop()
	RegisterResource(resourceType string, resourceName string, options config.Config) error
	//UnregisterResource(resource string) (i Interface, err error)
	HasResource(resource string) bool
	Resource(resource string) (r interface{}, status int)
	AddListener(l func(event int, ctx interface{}))
}

// Bootstrap is an abstract bootstrap struct
type Bootstrap struct {
	Options       *Config
	Resources     resource.Registry
	Services      service.Server
	Logger        *log.Log
	resourcesPool map[string]resource.Interface
	initialized   bool
	lsns          []func(event int, ctx interface{})
	mu            sync.Mutex
}

// Init initializes the application
func (b *Bootstrap) Init() error {
	cfg := b.Options.Get("resources")
	if cfg == nil {
		return errors.New("[Application] Resources configuration undefined")
	}

	if err := b.Resources.Init(cfg); err != nil {
		return err
	}
	b.Resources.Listen(b.throw)

	cfg = b.Options.Get("services")
	if cfg == nil {
		return errors.New("[Application] Services configuration undefined")
	}

	if err := b.Services.Init(cfg); err != nil {
		return err
	}
	b.Services.Listen(b.throw)

	b.initialized = true
	return nil
}

// Run Serves the application
func (b *Bootstrap) Run(ctx context.Context) error {
	return b.Services.Serve(ctx)
}

// Stop stops the application
func (b *Bootstrap) Stop() {
	b.Services.Stop()
}

// SetOptions Sets configuration for bootsrap struct
func (b *Bootstrap) SetOptions(options *Config) error {
	if options == nil {
		return nil
	}

	b.Options = options
	return nil
}

// GetOptions returns configuration of the bootstrap struct
func (b *Bootstrap) GetOptions() *Config {
	return b.Options
}

//RegisterResource registers a resource for bootstraping
func (b *Bootstrap) RegisterResource(resourceType string, resourceName string, options config.Config) error {
	if b.initialized {
		return errors.New("Registering resources is only allowed before initialization fase. Bootstrap already initialized")
	}

	if _, ok := b.resourcesPool[resourceName]; ok {
		return errors.Errorf("Resource by name '%s' is already registered.", resourceName)
	}

	rsr, err := resource.NewResource(resourceType, options)
	if err != nil {
		return err
	}

	b.Resources.Register(resourceName, resourceType, rsr)
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

// throw invokes event handler if any
func (b *Bootstrap) throw(event int, ctx interface{}) {
	for _, l := range b.lsns {
		l(event, ctx)
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
