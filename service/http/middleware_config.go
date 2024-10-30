package http

import (
	"github.com/noxyicm/wsf/config"
)

// Config defines HTTP server configuration
type MiddlewareConfig struct {
	Enable  bool
	Type    string
	Params  map[string]interface{}
	Headers map[string]string
}

// Populate populates Config values using given Config source
func (c *MiddlewareConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *MiddlewareConfig) Defaults() error {
	c.Enable = true
	c.Params = make(map[string]interface{})
	c.Headers = make(map[string]string)
	return nil
}

// Valid validates the configuration
func (c *MiddlewareConfig) Valid() error {
	return nil
}
