package router

import "wsf/config"

// Config represents router configuration
type Config struct {
	Type string
	File string
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
	c.File = "routes.json"
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
