package db

import (
	"wsf/config"
)

// Config defines set of adapter variables
type Config struct {
	Priority       int
	Adapter        string
	DefaultAdapter string
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(c)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Priority = 10
	return nil
}
