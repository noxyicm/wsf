package auth

import (
	"wsf/config"
)

// Config defines Static server configuration
type Config struct {
	Enable bool
	Type   string
	Header string
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Enable = true
	c.Type = TYPENone
	c.Header = "Authorization"
	return nil
}

// Valid returns nil if config is valid
func (c *Config) Valid() error {
	return nil
}
