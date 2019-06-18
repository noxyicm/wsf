package placeholder

import (
	"wsf/registry"
	"wsf/view/helper/placeholder/container"
)

// RegistryKey is a key by wich it will be stored in global registry
const RegistryKey = "WSFViewHalperPlaceholderRegistry"

// Registry is a placeholder registry
type Registry struct {
	items map[string]container.Interface
}

// CreateContainer returns new placeholder container
func (r *Registry) CreateContainer(key string, value []interface{}) (ci container.Interface, err error) {
	r.items[key], err = container.NewContainer(value)
	if err != nil {
		return nil, err
	}

	return r.items[key], nil
}

// GetContainer returns container by its key
func (r *Registry) GetContainer(key string) container.Interface {
	if v, ok := r.items[key]; ok {
		return v
	}

	container, _ := r.CreateContainer(key, []interface{}{})
	return container
}

// GetRegistry returns registered or creates new one registry
func GetRegistry() *Registry {
	if registry.Has(RegistryKey) {
		return registry.Get(RegistryKey).(*Registry)
	}

	rgs := &Registry{
		items: make(map[string]container.Interface),
	}

	registry.Set(RegistryKey, rgs)
	return rgs
}
