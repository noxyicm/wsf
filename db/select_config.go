package db

import "github.com/noxyicm/wsf/config"

// SelectConfig defines set of select variables
type SelectConfig struct {
	Type string
}

// Populate populates Config values using given Config source
func (c *SelectConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *SelectConfig) Defaults() error {
	c.Type = "default"
	return nil
}

// Valid validates the configuration
func (c *SelectConfig) Valid() error {
	return nil
}
