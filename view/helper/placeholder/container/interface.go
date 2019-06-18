package container

// Container contants
const (
	SET     = "SET"
	APPEND  = "APPEND"
	PREPEND = "PREPEND"
)

// Interface is a placeholder container interface
type Interface interface {
	Set(value interface{}) error
	Get() []interface{}
	GetOffset(index int) interface{}
	Unset(index int)
	Append(value interface{}) error
	Prepend(value interface{}) error
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
	data      []interface{}
	prefix    string
	postfix   string
	separator string
	indent    string
}

// Set sets a single value
func (c *Container) Set(value interface{}) error {
	c.data = []interface{}{value}
	return nil
}

// Get returns a container contents
func (c *Container) Get() []interface{} {
	return c.data
}

// GetOffset returns a container contents in a specific index
func (c *Container) GetOffset(index int) interface{} {
	if len(c.data) > index {
		return c.data[index]
	}

	return nil
}

// Unset specific index from container
func (c *Container) Unset(index int) {
	if len(c.data) > index {
		c.data = append(c.data[0:index-1], c.data[index:]...)
	}
}

// Append value to the bottom of the container
func (c *Container) Append(value interface{}) error {
	c.data = append(c.data, value)
	return nil
}

// Prepend value to the top of the container
func (c *Container) Prepend(value interface{}) error {
	c.data = append([]interface{}{value}, c.data...)
	return nil
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
func NewContainer(data []interface{}) (Interface, error) {
	return &Container{
		data: data,
	}, nil
}
