package modules

import "github.com/noxyicm/wsf/config"

// Config represents module configuration
type Config struct {
	Type     string
	Priority int
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
	c.Priority = 4
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
