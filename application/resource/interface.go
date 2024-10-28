package resource

import (
	"sync"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
)

const (
	// StatusUndefined when resource bus can not find the resource
	StatusUndefined = iota

	// StatusRegistered when resource has been registered in registry
	StatusRegistered

	// StatusInit when resource has been properly configured
	StatusInit

	// StatusOK when resource has been properly configured and setuped
	StatusOK

	// StatusStopped when resource stopped
	StatusStopped
)

var (
	buildHandlers = map[string]func(config.Config) (Interface, error){}
)

// Interface is a resource interface
type Interface interface {
	Priority() int
}

type bus struct {
	name     string
	typ      string
	resource Interface
	mu       sync.Mutex
	status   int
	order    int
}

// getStatus returns resource bus status
func (b *bus) getStatus() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.status
}

// setStatus sets resource bus in a specific status
func (b *bus) setStatus(status int) {
	b.mu.Lock()
	b.status = status
	b.mu.Unlock()
}

// hasStatus checks if resource bus in a specific status
func (b *bus) hasStatus(status int) bool {
	return b.getStatus() == status
}

// NewResource creates a new typed resource
func NewResource(resourceType string, cfg config.Config) (Interface, error) {
	if f, ok := buildHandlers[resourceType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized resource type \"%v\"", resourceType)
}

// Register registers a handler for resource creation
func Register(resourceType string, handler func(config.Config) (Interface, error)) {
	buildHandlers[resourceType] = handler
}
