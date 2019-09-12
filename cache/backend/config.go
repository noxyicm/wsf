package backend

import (
	"wsf/config"
)

// Config represents dispatcher configuration
type Config struct {
	Type          string
	Lifetime      int64
	Serialization bool
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
	c.Type = "default"
	c.Serialization = true
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
