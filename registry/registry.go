package registry

import (
	"sync"
)

var container *Container
var resources *Container

func init() {
	container = &Container{
		data: make(map[string]interface{}),
	}
	resources = &Container{
		data: make(map[string]interface{}),
	}
}

// Container represents maped values register
type Container struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// Get returns registered value
func (c *Container) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.data[key]; ok {
		return v
	}

	return nil
}

// Set sets new or resets old value
func (c *Container) Set(key string, value interface{}) {
	c.mu.Lock()
	c.data[key] = value
	c.mu.Unlock()
}

// Has return true if key registered in container
func (c *Container) Has(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[key]; ok {
		return true
	}

	return false
}

// Get returns registered value
func Get(key string) interface{} {
	return container.Get(key)
}

// GetBool returns registered value as bool
func GetBool(key string) bool {
	if v, ok := container.Get(key).(bool); ok {
		return v
	}

	return false
}

// GetInt returns registered value as int
func GetInt(key string) int {
	if v, ok := container.Get(key).(int); ok {
		return v
	}

	return 0
}

// GetString returns registered value as string
func GetString(key string) string {
	if v, ok := container.Get(key).(string); ok {
		return v
	}

	return ""
}

// Set sets new or resets old value
func Set(key string, value interface{}) {
	container.Set(key, value)
}

// Has return true if key registered in container
func Has(key string) bool {
	return container.Has(key)
}

// GetResource returns registered resource
func GetResource(name string) interface{} {
	return resources.Get(name)
}

// SetResource sets a ne resource
func SetResource(name string, value interface{}) {
	resources.Set(name, value)
}
