package db

import (
	"database/sql"
	"github.com/noxyicm/wsf/config"
)

// TransactionConfig represents transaction configuration
type TransactionConfig struct {
	Type           string
	IsolationLevel sql.IsolationLevel
	ReadOnly       bool
}

// Populate populates Config values using given Config source
func (c *TransactionConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *TransactionConfig) Defaults() error {
	c.Type = "default"
	c.IsolationLevel = sql.LevelDefault
	c.ReadOnly = false
	return nil
}

// Valid validates the configuration
func (c *TransactionConfig) Valid() error {
	return nil
}
