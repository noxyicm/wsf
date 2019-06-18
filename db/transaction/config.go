package transaction

import (
	"database/sql"
	"wsf/config"
)

// Config represents transaction configuration
type Config struct {
	Type           string
	IsolationLevel sql.IsolationLevel
	ReadOnly       bool
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
	c.IsolationLevel = sql.LevelDefault
	c.ReadOnly = false
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
