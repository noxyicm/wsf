package service

import (
	"sync"
	"wsf/config"
	"wsf/service"

	"github.com/pkg/errors"
)

const (
	// StatusUndefined when service bus can not find the service
	StatusUndefined = iota

	// StatusRegistered when service has been registered in server
	StatusRegistered

	// StatusOK when service has been properly configured
	StatusOK

	// StatusServing when service is serving
	StatusServing

	// StatusStopping when service in shutdown
	StatusStopping

	// StatusStopped when service stopped
	StatusStopped
)

var (
	buildHandlers = map[string]func(config.Config) (service.Interface, error){}
)

// service bus for underlying service
type bus struct {
	name    string
	typ     string
	service service.Interface
	mu      sync.Mutex
	status  int
	order   int
}

// getStatus returns service bus status
func (b *bus) getStatus() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.status
}

// setStatus sets service bus in a specific status
func (b *bus) setStatus(status int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
}

// hasStatus checks if service bus in a specific status
func (b *bus) hasStatus(status int) bool {
	return b.getStatus() == status
}

// canServe returns true if service can serve
func (b *bus) canServe() bool {
	_, ok := b.service.(service.Interface)
	return ok
}

// NewService creates a new typed service
func NewService(serviceType string, cfg config.Config) (service.Interface, error) {
	if f, ok := buildHandlers[serviceType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized service type \"%v\"", serviceType)
}

// Register registers a handler for service creation
func Register(serviceType string, handler func(config.Config) (service.Interface, error)) {
	buildHandlers[serviceType] = handler
}
