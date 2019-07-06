package backend

import (
	"sync"
	"wsf/config"
	"wsf/errors"
)

const (
	// TYPEBackend represents default backend cache
	TYPEBackend = "backend"
)

var (
	buildHandlers = map[string]func(config.Config) (Interface, error){}
)

func init() {
	Register(TYPEBackend, NewDefaultBackendCache)
}

// Interface represents backend cache interface
type Interface interface {
	Load(id string, testCacheValidity bool) ([]byte, error)
	Test(id string) bool
	Save(data []byte, id string, tags []string, specificLifetime int64) error
	Remove(id string) error
	Clear(mode int64, tags []string) error
}

// NewBackendCache creates a new backend cache specified by type
func NewBackendCache(backendType string, options config.Config) (Interface, error) {
	if f, ok := buildHandlers[backendType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized backend cache type \"%v\"", backendType)
}

// Register registers a handler for backend cache creation
func Register(backendType string, handler func(config.Config) (Interface, error)) {
	buildHandlers[backendType] = handler
}

// Backend chache handler
type Backend struct {
	Options *Config
	mu      sync.Mutex
}

// Load stored data
func (b *Backend) Load(id string, testCacheValidity bool) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return []byte{}, nil
}

// Test if key exists
func (b *Backend) Test(id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return false
}

// Save data by key
func (b *Backend) Save(data []byte, id string, tags []string, specificLifetime int64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return nil
}

// Remove data by key
func (b *Backend) Remove(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return nil
}

// Clean stored data by tags
func (b *Backend) Clear(mode int64, tags []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return nil
}

// NewDefaultBackendCache creates new default backend cache
func NewDefaultBackendCache(options config.Config) (Interface, error) {
	b := &Backend{}

	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)
	b.Options = cfg

	return b, nil
}
