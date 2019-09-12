package container

import (
	"wsf/utils/stack"
)

// Container contants
const (
	SET     = "SET"
	APPEND  = "APPEND"
	PREPEND = "PREPEND"
)

// Interface is a placeholder container interface
type Interface interface {
	Set(key string, value interface{}) error
	Get(key string) interface{}
	GetAll() map[string]interface{}
	GetStack() []interface{}
	Unset(key string)
	Append(key string, value interface{}) error
	Prepend(key string, value interface{}) error
	SetPrefix(prefix string) error
	Prefix() string
	SetPostfix(postfix string) error
	Postfix() string
	SetSeparator(separator string) error
	Separator() string
	SetIndent(indent string) error
	Indent() string
}

// Container is a default placeholder container implementation
type Container struct {
	data      *stack.Referenced
	prefix    string
	postfix   string
	separator string
	indent    string
}

// Set sets a single value
func (c *Container) Set(key string, value interface{}) error {
	c.data = stack.NewReferenced(nil)
	return c.data.Set(key, value)
}

// Get returns a container value
func (c *Container) Get(key string) interface{} {
	if c.data.Has(key) {
		return c.data.Value(key)
	}

	return nil
}

// GetAll returns a container contents
func (c *Container) GetAll() map[string]interface{} {
	return c.data.Map()
}

// GetStack returns a container contents as slice
func (c *Container) GetStack() []interface{} {
	return c.data.Stack()
}

// Unset specific index from container
func (c *Container) Unset(key string) {
	c.data.Unset(key)
}

// Append value to the bottom of the container
func (c *Container) Append(key string, value interface{}) error {
	return c.data.Append(key, value)
}

// Prepend value to the top of the container
func (c *Container) Prepend(key string, value interface{}) error {
	return c.data.Prepend(key, value)
}

// SetPrefix sets a prefix for serialization
func (c *Container) SetPrefix(prefix string) error {
	c.prefix = prefix
	return nil
}

// Prefix returns container prefix
func (c *Container) Prefix() string {
	return c.prefix
}

// SetPostfix sets a postfix for serialization
func (c *Container) SetPostfix(postfix string) error {
	c.postfix = postfix
	return nil
}

// Postfix returns container postfix
func (c *Container) Postfix() string {
	return c.postfix
}

// SetSeparator sets a separator for serialization
func (c *Container) SetSeparator(separator string) error {
	c.separator = separator
	return nil
}

// Separator returns container separator
func (c *Container) Separator() string {
	return c.separator
}

// SetIndent sets a indent for serialization
func (c *Container) SetIndent(indent string) error {
	c.indent = indent
	return nil
}

// Indent returns container indent
func (c *Container) Indent() string {
	return c.indent
}

// NewContainer creates a new placeholder container
func NewContainer(data map[string]interface{}) (Interface, error) {
	return &Container{
		data: stack.NewReferenced(data),
	}, nil
}
